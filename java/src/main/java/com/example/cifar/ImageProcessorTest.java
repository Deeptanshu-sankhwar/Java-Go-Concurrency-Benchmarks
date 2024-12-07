package com.example.cifar;

import org.junit.Test;

import java.io.File;
import java.io.FileWriter;
import java.io.PrintWriter;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.List;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.atomic.AtomicBoolean;

import static org.junit.Assert.*;

public class ImageProcessorTest {

    @Test
    public void testProduceBatches() {
        // Prepare mock data
        BlockingQueue<float[][]> queue = new LinkedBlockingQueue<>(10);
        AtomicBoolean isDoneProducing = new AtomicBoolean(false);
        float[][] mockBatch = new float[1][ImageProcessor.IMAGE_SIZE];
        for (int i = 0; i < ImageProcessor.IMAGE_SIZE; i++) {
            mockBatch[0][i] = i / 255.0f;
        }

        // Produce batches
        new Thread(() -> ImageProcessor.produceBatches(List.of(mockBatch, mockBatch), queue, isDoneProducing)).start();

        // Wait for producer to finish
        while (!isDoneProducing.get()) {
            // Spinlock for demonstration purposes
        }

        // Validate queue
        assertEquals(2, queue.size());
        assertTrue(isDoneProducing.get());
    }

    @Test
    public void testProcessBatch() {
        // Prepare mock batch
        float[][] mockBatch = new float[1][ImageProcessor.IMAGE_SIZE];
        for (int i = 0; i < ImageProcessor.IMAGE_SIZE; i++) {
            mockBatch[0][i] = i / 255.0f;
        }

        // Process the batch
        ImageProcessor.processBatch(mockBatch);

        // Validate the processed batch
        for (int i = 0; i < ImageProcessor.IMAGE_SIZE; i++) {
            assertEquals((i / 255.0f) * 2, mockBatch[0][i], 1e-6);
        }
    }

    @Test
    public void testLogging() throws Exception {
        String logFilePath = "test_cifar_metrics.log";
        String testMessage = "This is a test log message.";

        // Log a message
        ImageProcessor.logMessage(logFilePath, testMessage);

        // Validate the log file
        String logContent = Files.readString(Paths.get(logFilePath));
        assertTrue(logContent.contains(testMessage));

        // Clean up
        Files.deleteIfExists(Paths.get(logFilePath));
    }
}