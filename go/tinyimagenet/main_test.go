package main

import (
	"sync"
	"testing"
	"time"
)

func TestSimulateImageProcessing(t *testing.T) {
	image := make([]float32, imageHeight*imageWidth*channels)
	for i := range image {
		image[i] = 1.0
	}

	processedImage := SimulateImageProcessing(image)
	for i, val := range processedImage {
		if val != 2.0 {
			t.Errorf("Pixel %d value mismatch: expected 2.0, got %.2f", i, val)
		}
	}
}

func TestProcessBatch(t *testing.T) {
	batch := ImageBatch{
		Images: make([][]float32, batchSize),
		Labels: make([]string, batchSize),
	}

	for i := 0; i < batchSize; i++ {
		image := make([]float32, imageHeight*imageWidth*channels)
		for j := 0; j < imageHeight*imageWidth*channels; j++ {
			image[j] = 1.0
		}
		batch.Images[i] = image
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go ProcessBatch(batch, &wg)
	wg.Wait()

	for i, img := range batch.Images {
		for j, val := range img {
			if val != 2.0 {
				t.Errorf("Batch %d image %d pixel %d mismatch: expected 2.0, got %.2f", i, i, j, val)
			}
		}
	}
}

func TestRunProcessingTask(t *testing.T) {
	dataDir := "../../tiny-imagenet-200/train"
	images, labels, err := LoadTinyImageNet(dataDir)
	if err != nil {
		t.Fatalf("Failed to load Tiny ImageNet dataset: %v", err)
	}

	executionTime, concurrencyOverhead := RunProcessingTask(images, labels)
	if executionTime == 0 {
		t.Errorf("Execution time should not be zero")
	}
	if concurrencyOverhead < executionTime {
		t.Errorf("Concurrency overhead should be greater than or equal to execution time")
	}
}

func TestCalculateCPUUsage(t *testing.T) {
	duration := 2 * time.Second
	cpuUsage, err := calculateCPUUsage(duration)
	if err != nil {
		t.Fatalf("Failed to calculate CPU usage: %v", err)
	}

	if cpuUsage < 0 || cpuUsage > 100 {
		t.Errorf("CPU usage out of bounds: %.2f%%", cpuUsage)
	}
}
