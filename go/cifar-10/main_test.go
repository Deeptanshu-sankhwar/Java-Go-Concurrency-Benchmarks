package main

import (
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestLoadCIFAR10(t *testing.T) {
	dataDir := "../../cifar-10-batches-bin/"
	images, labels, err := LoadCIFAR10(dataDir)
	if err != nil {
		t.Fatalf("Failed to load CIFAR-10 dataset: %v", err)
	}

	if len(images) != 50000 {
		t.Errorf("Expected 50000 images, got %d", len(images))
	}

	if len(labels) != 50000 {
		t.Errorf("Expected 50000 labels, got %d", len(labels))
	}

	for i, img := range images {
		if len(img) != imageSize {
			t.Errorf("Image %d size mismatch: expected %d, got %d", i, imageSize, len(img))
		}
	}
}

func TestSimulateImageProcessing(t *testing.T) {
	image := make([]float32, imageSize)
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
		Labels: make([]int, batchSize),
	}

	for i := 0; i < batchSize; i++ {
		image := make([]float32, imageSize)
		for j := 0; j < imageSize; j++ {
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

func TestAppendToLogFile(t *testing.T) {
	logFilePath := "test_log.log"
	message := "Test log message"

	err := AppendToLogFile(logFilePath, message)
	if err != nil {
		t.Fatalf("Failed to append to log file: %v", err)
	}

	data, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(data), message) {
		t.Errorf("Log file content mismatch: expected message not found")
	}

	err = os.Remove(logFilePath)
	if err != nil {
		t.Fatalf("Failed to delete log file during cleanup: %v", err)
	}
}
func contains(data, substring string) bool {
	return len(data) >= len(substring) && data[:len(substring)] == substring
}

func TestRunProcessingTask(t *testing.T) {
	dataDir := "../../cifar-10-batches-bin/"
	images, labels, err := LoadCIFAR10(dataDir)
	if err != nil {
		t.Fatalf("Failed to load CIFAR-10 dataset: %v", err)
	}

	executionTime, concurrencyOverhead := RunProcessingTask(images, labels)
	if executionTime == 0 {
		t.Errorf("Execution time should not be zero")
	}
	if concurrencyOverhead < executionTime {
		t.Errorf("Concurrency overhead should be greater than or equal to execution time")
	}
}
