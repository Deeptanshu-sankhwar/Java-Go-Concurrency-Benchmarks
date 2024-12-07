package com.example.tinyimagenet;

import javax.imageio.ImageIO;
import java.awt.image.BufferedImage;
import java.io.FileWriter;
import java.io.IOException;
import java.io.PrintWriter;
import java.lang.management.ManagementFactory;
import com.sun.management.OperatingSystemMXBean;
import java.nio.file.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicBoolean;

public class ImageProcessor {

    // Configuration
    static final int BATCH_SIZE = 500;
    static final int NUM_RUNS = 100;
    static final int QUEUE_CAPACITY = 50;
    static final int NUM_CONSUMERS = Runtime.getRuntime().availableProcessors();
    static final String DATA_DIR = "../../tiny-imagenet-200/train";
    static final String LOG_FILE_PATH = "java_tinyimagenet_metrics_result.log";

    public static void main(String[] args) throws Exception {
        logMessage(LOG_FILE_PATH, "Loading Tiny ImageNet dataset with metrics monitoring...");

        long totalExecutionTime = 0;
        long totalMemoryUsage = 0;
        double totalCpuUsage = 0;
        long totalOverheadTime = 0;

        for (int run = 0; run < NUM_RUNS; run++) {
            logMessage(LOG_FILE_PATH, String.format("Run %d/%d...\n", run + 1, NUM_RUNS));

            BlockingQueue<BufferedImage> queue = new LinkedBlockingQueue<>(QUEUE_CAPACITY);
            AtomicBoolean isDoneProducing = new AtomicBoolean(false);

            // Producer thread
            Thread producer = new Thread(() -> produceImages(DATA_DIR, queue, isDoneProducing));

            Runtime runtime = Runtime.getRuntime();
            OperatingSystemMXBean osBean = ManagementFactory.getPlatformMXBean(OperatingSystemMXBean.class);

            runtime.gc();
            long memoryBefore = runtime.totalMemory() - runtime.freeMemory();

            long startOverheadTime = System.nanoTime();
            producer.start(); // Start producer thread
            long startTime = System.nanoTime();
            processImages(queue, isDoneProducing); // Process images with consumers
            long endTime = System.nanoTime();
            long endOverheadTime = System.nanoTime();

            producer.join(); // Ensure producer completes

            runtime.gc();
            long memoryAfter = runtime.totalMemory() - runtime.freeMemory();
            long memoryUsage = memoryAfter - memoryBefore;

            double cpuUsage = osBean.getProcessCpuLoad();

            long executionTime = endTime - startTime;
            long overheadTime = endOverheadTime - startOverheadTime;

            totalExecutionTime += executionTime;
            totalMemoryUsage += memoryUsage;
            totalCpuUsage += cpuUsage;
            totalOverheadTime += overheadTime;

            logMessage(LOG_FILE_PATH, String.format("Execution Time for Run %d: %.2f seconds", run + 1, executionTime / 1e9));
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

    // Producer method to enqueue images
    private static void produceImages(String directory, BlockingQueue<BufferedImage> queue, AtomicBoolean isDoneProducing) {
        try {
            Files.walk(Paths.get(directory)).forEach(path -> {
                try {
                    if (Files.isRegularFile(path) && path.toString().endsWith(".jpg")) {
                        BufferedImage img = ImageIO.read(path.toFile());
                        queue.put(img); // Blocks if queue is full
                    }
                } catch (Exception e) {
                    logMessage(LOG_FILE_PATH, "Error processing image: " + e.getMessage());
                }
            });
        } catch (Exception e) {
            logMessage(LOG_FILE_PATH, "Error walking dataset directory: " + e.getMessage());
        } finally {
            isDoneProducing.set(true);
        }
    }

    // Consumer pool processes images from the queue
    private static void processImages(BlockingQueue<BufferedImage> queue, AtomicBoolean isDoneProducing) {
        ExecutorService consumers = Executors.newFixedThreadPool(NUM_CONSUMERS);

        Runnable consumerTask = () -> {
            try {
                while (!isDoneProducing.get() || !queue.isEmpty()) {
                    BufferedImage image = queue.poll(1, TimeUnit.SECONDS);
                    if (image != null) {
                        processImage(image);
                    }
                }
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                logMessage(LOG_FILE_PATH, "Consumer thread interrupted: " + e.getMessage());
            }
        };

        for (int i = 0; i < NUM_CONSUMERS; i++) {
            consumers.execute(consumerTask);
        }

        consumers.shutdown();
        try {
            consumers.awaitTermination(10, TimeUnit.MINUTES);
        } catch (InterruptedException e) {
            logMessage(LOG_FILE_PATH, "Consumer threads failed to terminate in time.");
        }
    }

    // Simulated processing (e.g., pixel value multiplication)
    private static void processImage(BufferedImage image) {
        for (int y = 0; y < image.getHeight(); y++) {
            for (int x = 0; x < image.getWidth(); x++) {
                int rgb = image.getRGB(x, y);
                int r = (rgb >> 16) & 0xFF;
                int g = (rgb >> 8) & 0xFF;
                int b = rgb & 0xFF;

                r = Math.min(r * 2, 255);
                g = Math.min(g * 2, 255);
                b = Math.min(b * 2, 255);

                int newRgb = (r << 16) | (g << 8) | b;
                image.setRGB(x, y, newRgb);
            }
        }
    }

    // Log message to file
    public static void logMessage(String filePath, String message) {
        try (PrintWriter out = new PrintWriter(new FileWriter(filePath, true))) {
            out.println(message);
            System.out.println(message);
        } catch (IOException e) {
            System.err.println("Failed to write log: " + e.getMessage());
        }
    }
}

