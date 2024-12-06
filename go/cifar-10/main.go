package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

const (
	numImagesPerBatch = 10000
	imageHeight       = 32
	imageWidth        = 32
	channels          = 3
	imageSize         = imageHeight * imageWidth * channels
	batchSize         = 500 // Processing batch size
	numRuns           = 100 // Number of times to repeat the task for averaging
)

// ImageBatch represents a batch of images
type ImageBatch struct {
	Images [][]float32
	Labels []int
}

// LoadCIFAR10 loads all CIFAR-10 dataset batches
func LoadCIFAR10(dataDir string) ([][]float32, []int, error) {
	var allImages [][]float32
	var allLabels []int

	// Load all 5 batches
	for i := 1; i <= 5; i++ {
		filePath := filepath.Join(dataDir, fmt.Sprintf("data_batch_%d.bin", i))
		fmt.Printf("Loading batch: %s\n", filePath)

		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
		}

		for j := 0; j < numImagesPerBatch; j++ {
			label := int(data[j*(imageSize+1)])
			imageData := data[j*(imageSize+1)+1 : (j+1)*(imageSize+1)]
			image := make([]float32, imageSize)
			for k := 0; k < imageSize; k++ {
				image[k] = float32(imageData[k]) / 255.0
			}

			allImages = append(allImages, image)
			allLabels = append(allLabels, label)
		}
	}
	return allImages, allLabels, nil
}

// SimulateImageProcessing performs dummy image transformations
func SimulateImageProcessing(image []float32) []float32 {
	for i := range image {
		image[i] = image[i] * 2
	}
	return image
}

// ProcessBatch processes a batch of images concurrently
func ProcessBatch(batch ImageBatch, wg *sync.WaitGroup) {
	defer wg.Done()
	for i, image := range batch.Images {
		batch.Images[i] = SimulateImageProcessing(image)
	}
}

// RunProcessingTask runs the preprocessing task once and returns execution time and concurrency overhead
func RunProcessingTask(images [][]float32, labels []int) (time.Duration, time.Duration) {
	// Divide into batches
	totalImages := len(images)
	numBatches := totalImages / batchSize
	batches := make([]ImageBatch, numBatches)
	for i := 0; i < numBatches; i++ {
		start := i * batchSize
		end := start + batchSize
		batches[i] = ImageBatch{
			Images: images[start:end],
			Labels: labels[start:end],
		}
	}

	// Start concurrent processing
	startOverhead := time.Now()
	startExecution := time.Now()

	var wg sync.WaitGroup
	for _, batch := range batches {
		wg.Add(1)
		go ProcessBatch(batch, &wg)
	}
	wg.Wait()

	executionTime := time.Since(startExecution)
	concurrencyOverhead := time.Since(startOverhead)
	return executionTime, concurrencyOverhead
}

// AppendToLogFile appends a string to the specified log file
func AppendToLogFile(filePath, message string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	logger := log.New(file, "", log.LstdFlags)
	logger.Println(message)

	return nil
}

// calculateCPUUsage calculates average CPU utilization during a processing window
func calculateCPUUsage(duration time.Duration) (float64, error) {
	percentages, err := cpu.Percent(duration, false)
	if err != nil {
		return 0, err
	}
	return percentages[0], nil
}

func main() {
	logFilePath := "go_cifar10_metrics_result.log"

	// Load CIFAR-10 dataset
	err := AppendToLogFile(logFilePath, "Loading CIFAR-10 dataset...")
	dataDir := "../../cifar-10-batches-bin/"
	images, labels, err := LoadCIFAR10(dataDir)
	if err != nil {
		log.Fatalf("Error loading CIFAR-10: %v", err)
	}
	err = AppendToLogFile(logFilePath, "Dataset loaded successfully.")

	err = AppendToLogFile(logFilePath, "\nDataset Parameters:")
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Total Images: %d\n", len(images)))
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Image Shape: %d x %d x %d (Height x Width x Channels)\n", imageHeight, imageWidth, channels))
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Number of Classes: %d\n", 10))

	var totalExecutionTime, totalConcurrencyOverhead time.Duration
	var totalMemoryUsage uint64
	var totalCPUUsage float64

	for i := 0; i < numRuns; i++ {
		err = AppendToLogFile(logFilePath, fmt.Sprintf("\nRun %d/%d...\n", i+1, numRuns))

		var memStatsBefore runtime.MemStats
		runtime.ReadMemStats(&memStatsBefore)
		memoryBefore := memStatsBefore.Alloc

		executionTime, concurrencyOverhead := RunProcessingTask(images, labels)

		var memStatsAfter runtime.MemStats
		runtime.ReadMemStats(&memStatsAfter)
		memoryAfter := memStatsAfter.Alloc
		memoryUsage := memoryAfter - memoryBefore

		startCPUTime := time.Now()
		cpuUsage, err := calculateCPUUsage(time.Since(startCPUTime))
		if err != nil {
			log.Fatalf("Error calculating CPU usage: %v", err)
		}

		totalExecutionTime += executionTime
		totalConcurrencyOverhead += concurrencyOverhead
		totalMemoryUsage += memoryUsage
		totalCPUUsage += cpuUsage

		err = AppendToLogFile(logFilePath, fmt.Sprintf("Execution Time for Run %d: %.2f seconds", i+1, executionTime.Seconds()))
		err = AppendToLogFile(logFilePath, fmt.Sprintf("Concurrency Overhead for Run %d: %.2f seconds", i+1, concurrencyOverhead.Seconds()))
		err = AppendToLogFile(logFilePath, fmt.Sprintf("Memory Usage for Run %d: %.2f MB", i+1, float64(memoryUsage)/(1024*1024)))
		err = AppendToLogFile(logFilePath, fmt.Sprintf("CPU Utilization for Run %d: %.2f%%", i+1, cpuUsage*100))
	}

	err = AppendToLogFile(logFilePath, "\nAverage Metrics:")
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Average Execution Time: %.2f seconds", totalExecutionTime.Seconds()/float64(numRuns)))
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Average Concurrency Overhead: %.2f seconds", totalConcurrencyOverhead.Seconds()/float64(numRuns)))
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Average Memory Usage: %.2f MB", float64(totalMemoryUsage)/(float64(numRuns)*1024*1024)))
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Average CPU Utilization: %.2f%%", (totalCPUUsage/float64(numRuns))*100))
}
