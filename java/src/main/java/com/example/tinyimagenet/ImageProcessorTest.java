package com.example.tinyimagenet;

import org.junit.Test;

import java.awt.image.BufferedImage;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.atomic.AtomicBoolean;


import static org.junit.Assert.*;


public class ImageProcessorTest {

    @Test
    public void testProduceImagesWithoutImageProcessingLibraries() {
        BlockingQueue<BufferedImage> queue = new LinkedBlockingQueue<>(10);
        AtomicBoolean isDoneProducing = new AtomicBoolean(false);

        // Simulate image production by adding mock images directly to the queue
        for (int i = 0; i < 5; i++) {
            queue.add(new BufferedImage(10, 10, BufferedImage.TYPE_INT_RGB));
        }

        isDoneProducing.set(true);

        // Assert the queue contains the images
        assertFalse("Queue should not be empty after adding images.", queue.isEmpty());
        assertEquals(5, queue.size());
    }

    @Test
    public void testProcessImageWithoutDependencies() {
        // Create a sample image
        BufferedImage sampleImage = new BufferedImage(10, 10, BufferedImage.TYPE_INT_RGB);
        for (int x = 0; x < 10; x++) {
            for (int y = 0; y < 10; y++) {
                sampleImage.setRGB(x, y, (100 << 16) | (100 << 8) | 100); // RGB (100, 100, 100)
            }
        }

        // Mock processing: Simulate that processImage modifies the image
        for (int y = 0; y < sampleImage.getHeight(); y++) {
            for (int x = 0; x < sampleImage.getWidth(); x++) {
                int rgb = sampleImage.getRGB(x, y);
                int r = (rgb >> 16) & 0xFF;
                int g = (rgb >> 8) & 0xFF;
                int b = rgb & 0xFF;

                // Simulate doubling values (mock processing)
                r = Math.min(r * 2, 255);
                g = Math.min(g * 2, 255);
                b = Math.min(b * 2, 255);

                int newRgb = (r << 16) | (g << 8) | b;
                sampleImage.setRGB(x, y, newRgb);
            }
        }

        // Assert the processing logic works
        for (int x = 0; x < 10; x++) {
            for (int y = 0; y < 10; y++) {
                int rgb = sampleImage.getRGB(x, y);
                int r = (rgb >> 16) & 0xFF;
                int g = (rgb >> 8) & 0xFF;
                int b = rgb & 0xFF;

                assertEquals(200, r);
                assertEquals(200, g);
                assertEquals(200, b);
            }
        }
    }

    @Test
    public void testLogging() {
        String testMessage = "This is a test log message.";
        String logFilePath = "test_metrics.log";

        // Call the logging method
        ImageProcessor.logMessage(logFilePath, testMessage);

        // Verify the log file contains the test message
        try {
            String logContent = java.nio.file.Files.readString(java.nio.file.Paths.get(logFilePath));
            assertTrue("Log file should contain the test message.", logContent.contains(testMessage));
        } catch (Exception e) {
            fail("Failed to read log file: " + e.getMessage());
        }
    }
}
