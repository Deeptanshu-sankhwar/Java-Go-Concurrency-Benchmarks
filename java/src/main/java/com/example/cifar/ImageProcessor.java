package com.example.cifar;

import java.io.*;
import java.lang.management.ManagementFactory;
import com.sun.management.OperatingSystemMXBean;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicBoolean;

public class ImageProcessor {

    // Configuration
    static final int IMAGE_HEIGHT = 32;
    static final int IMAGE_WIDTH = 32;
    static final int CHANNELS = 3;
    static final int IMAGE_SIZE = IMAGE_HEIGHT * IMAGE_WIDTH * CHANNELS;
    static final int BATCH_SIZE = 500;
    static final int NUM_IMAGES_PER_BATCH = 10000;
    static final int NUM_RUNS = 100;
    static final int QUEUE_CAPACITY = 10;
    static final int NUM_CONSUMERS = Runtime.getRuntime().availableProcessors();
    static final String LOG_FILE_PATH = "java_cifar10_metrics_result.log";

    public static void main(String[] args) throws Exception {
        String dataDir = "src/main/java/com/example/cifar/cifar-10-batches-bin/";
        logMessage(LOG_FILE_PATH, "Loading CIFAR-10 dataset into memory...");

        List<float[][]> allBatches = loadCifar10Dataset(dataDir);

        logMessage(LOG_FILE_PATH, "Dataset loaded successfully.");
        long totalExecutionTime = 0;
        long totalMemoryUsage = 0;
        double totalCpuUsage = 0;
        long totalOverheadTime = 0;

        OperatingSystemMXBean osBean = ManagementFactory.getPlatformMXBean(OperatingSystemMXBean.class);

        for (int run = 0; run < NUM_RUNS; run++) {
            logMessage(LOG_FILE_PATH, String.format("Run %d/%d...\n", run + 1, NUM_RUNS));

            BlockingQueue<float[][]> queue = new LinkedBlockingQueue<>(QUEUE_CAPACITY);
            AtomicBoolean isDoneProducing = new AtomicBoolean(false);

            Runtime runtime = Runtime.getRuntime();
            runtime.gc();
            long memoryBefore = runtime.totalMemory() - runtime.freeMemory();

            // Producer thread reuses pre-loaded batches
            Thread producer = new Thread(() -> produceBatches(allBatches, queue, isDoneProducing));
            producer.start();

            long startOverheadTime = System.nanoTime();

            long startTime = System.nanoTime();
            processBatches(queue, isDoneProducing);
            long elapsedTime = System.nanoTime() - startTime;
            long endOverheadTime = System.nanoTime();

            producer.join();

            runtime.gc();
            long memoryAfter = runtime.totalMemory() - runtime.freeMemory();
            long memoryUsage = memoryAfter - memoryBefore;

            long overheadTime = endOverheadTime - startOverheadTime;

            double cpuUsage = osBean.getProcessCpuLoad();

            totalExecutionTime += elapsedTime;
            totalMemoryUsage += memoryUsage;
            totalCpuUsage += cpuUsage;
            totalOverheadTime += overheadTime;

            logMessage(LOG_FILE_PATH, String.format("Execution Time for Run %d: %.2f seconds", run + 1, elapsedTime / 1e9));
            logMessage(LOG_FILE_PATH, String.format("Memory Usage for Run %d: %.2f MB", run + 1, memoryUsage / (1024.0 * 1024.0)));
            logMessage(LOG_FILE_PATH, String.format("CPU Utilization for Run %d: %.2f%%", run + 1, cpuUsage * 100));
            logMessage(LOG_FILE_PATH, String.format("Concurrency Overhead for Run %d: %.2f seconds", run + 1, overheadTime / 1e9));
        }

        double averageExecutionTime = totalExecutionTime / (double) NUM_RUNS;
        double averageMemoryUsage = totalMemoryUsage / (double) NUM_RUNS;
        double averageCpuUsage = totalCpuUsage / NUM_RUNS;
        double averageOverheadTime = totalOverheadTime / (double) NUM_RUNS;

        logMessage(LOG_FILE_PATH, "\nAverage Metrics Across Runs:");
        logMessage(LOG_FILE_PATH, String.format("Average Execution Time: %.2f seconds", averageExecutionTime / 1e9));
        logMessage(LOG_FILE_PATH, String.format("Average Memory Usage: %.2f MB", averageMemoryUsage / (1024.0 * 1024.0)));
        logMessage(LOG_FILE_PATH, String.format("Average CPU Utilization: %.2f%%", averageCpuUsage * 100));
        logMessage(LOG_FILE_PATH, String.format("Average Concurrency Overhead: %.2f seconds", averageOverheadTime / 1e9));
    }

    static List<float[][]> loadCifar10Dataset(String dataDir) throws IOException {
        List<float[][]> allBatches = new ArrayList<>();
        for (int i = 1; i <= 5; i++) {
            String filePath = dataDir + String.format("data_batch_%d.bin", i);
            logMessage(LOG_FILE_PATH, String.format("Loading batch: %s\n", filePath));

            File file = new File(filePath);
            long expectedSize = NUM_IMAGES_PER_BATCH * (1 + IMAGE_SIZE); // 1 byte for label + IMAGE_SIZE bytes for image
            if (file.length() < expectedSize) {
                throw new IOException(String.format("File %s is smaller than expected. Size: %d bytes, Expected: %d bytes",
                        filePath, file.length(), expectedSize));
            }

            try (DataInputStream dis = new DataInputStream(new FileInputStream(file))) {
                for (int j = 0; j < NUM_IMAGES_PER_BATCH; j++) {
                    if (dis.available() < (1 + IMAGE_SIZE)) {
                        throw new EOFException(String.format("Unexpected end of file in %s while reading image %d", filePath, j));
                    }
                    dis.readByte(); // Read label (1 byte)
                    float[] image = new float[IMAGE_SIZE];
                    for (int k = 0; k < IMAGE_SIZE; k++) {
                        image[k] = (dis.readUnsignedByte() / 255.0f);
                    }
                    allBatches.add(new float[][]{image});
                }
            }
        }
        return allBatches;
    }

    static void produceBatches(List<float[][]> allBatches, BlockingQueue<float[][]> queue, AtomicBoolean isDoneProducing) {
        try {
            for (float[][] batch : allBatches) {
                queue.put(batch);
            }
        } catch (Exception e) {
            e.printStackTrace();
        } finally {
            isDoneProducing.set(true);
        }
    }

    private static void processBatches(BlockingQueue<float[][]> queue, AtomicBoolean isDoneProducing) {
        ExecutorService consumers = Executors.newFixedThreadPool(NUM_CONSUMERS);

        Runnable consumerTask = () -> {
            try {
                while (!isDoneProducing.get() || !queue.isEmpty()) {
                    float[][] batch = queue.poll(1, TimeUnit.SECONDS);
                    if (batch != null) {
                        processBatch(batch);
                    }
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
            }
        };

        for (int i = 0; i < NUM_CONSUMERS; i++) {
            consumers.execute(consumerTask);
        }

        consumers.shutdown();
        try {
            consumers.awaitTermination(10, TimeUnit.MINUTES);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
    }

    static void processBatch(float[][] batch) {
        for (float[] image : batch) {
            for (int i = 0; i < image.length; i++) {
                image[i] *= 2;
            }
        }
    }

    static void logMessage(String filePath, String message) {
        try (PrintWriter out = new PrintWriter(new FileWriter(filePath, true))) {
            out.println(message);
            System.out.println(message);
        } catch (IOException e) {
            System.err.println("Failed to write log: " + e.getMessage());
        }
    }
}
