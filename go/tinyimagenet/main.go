package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	_ "image/png"

	"github.com/shirou/gopsutil/cpu"
)

const (
	imageHeight = 64
	imageWidth  = 64
	channels    = 3
	batchSize   = 500 // Processing batch size
	numRuns     = 100 // Number of times to repeat the task for averaging
)

// ImageBatch represents a batch of images
type ImageBatch struct {
	Images [][]float32
	Labels []string
}

// LoadTinyImageNet loads all images and their labels from a specified directory
func LoadTinyImageNet(dataDir string) ([][]float32, []string, error) {
	var allImages [][]float32
	var allLabels []string

	fmt.Println("Loading Tiny ImageNet dataset...")

	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (filepath.Ext(path) == ".jpg" || filepath.Ext(path) == ".png") {
			img, label, err := loadImage(path)
			if err != nil {
				return fmt.Errorf("failed to load image %s: %v", path, err)
			}
			allImages = append(allImages, img)
			allLabels = append(allLabels, label)
		}
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to walk through dataset directory: %v", err)
	}

	return allImages, allLabels, nil
}

// loadImage loads and preprocesses a single image
func loadImage(imagePath string) ([]float32, string, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open image: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %v", err)
	}

	pixels := make([]float32, imageHeight*imageWidth*channels)
	idx := 0
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			pixels[idx] = float32(r) / 65535.0
			pixels[idx+1] = float32(g) / 65535.0
			pixels[idx+2] = float32(b) / 65535.0
			idx += 3
		}
	}

	label := filepath.Base(filepath.Dir(imagePath))
	return pixels, label, nil
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
func RunProcessingTask(images [][]float32, labels []string) (time.Duration, time.Duration) {
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
	percentages, err := cpu.Percent(duration, false) // Measure CPU usage over the given duration
	if err != nil {
		return 0, err
	}
	return percentages[0], nil
}

// Main function
func main() {
	logFilePath := "go_tinyimagenet_metrics_result.log"

	// Load Tiny ImageNet dataset
	dataDir := "../../tiny-imagenet-200/train"
	images, labels, err := LoadTinyImageNet(dataDir)
	if err != nil {
		log.Fatalf("Error loading Tiny ImageNet: %v", err)
	}
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Dataset loaded successfully. Total Images: %d\n", len(images)))

	err = AppendToLogFile(logFilePath, "\nDataset Parameters:")
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Total Images: %d\n", len(images)))
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Image Shape: %d x %d x %d (Height x Width x Channels)\n", imageHeight, imageWidth, channels))
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Number of Classes: %d\n", len(labels)))

	var totalExecutionTime, totalConcurrencyOverhead time.Duration
	var totalMemoryUsage uint64
	var totalCPUUsage float64

	for i := 0; i < numRuns; i++ {
		err = AppendToLogFile(logFilePath, fmt.Sprintf("\nRun %d/%d...\n", i+1, numRuns))

		var memStatsBefore runtime.MemStats
		runtime.ReadMemStats(&memStatsBefore)
		memoryBefore := memStatsBefore.Alloc

		startCPUTime := time.Now()
		executionTime, concurrencyOverhead := RunProcessingTask(images, labels)
		cpuUsage, err := calculateCPUUsage(time.Since(startCPUTime))
		if err != nil {
			log.Fatalf("Error calculating CPU usage: %v", err)
		}

		var memStatsAfter runtime.MemStats
		runtime.ReadMemStats(&memStatsAfter)
		memoryAfter := memStatsAfter.Alloc
		memoryUsage := memoryAfter - memoryBefore

		totalExecutionTime += executionTime
		totalConcurrencyOverhead += concurrencyOverhead
		totalMemoryUsage += memoryUsage
		totalCPUUsage += cpuUsage

		err = AppendToLogFile(logFilePath, fmt.Sprintf("Execution Time for Run %d: %.9f seconds", i+1, executionTime.Seconds()))
		err = AppendToLogFile(logFilePath, fmt.Sprintf("Concurrency Overhead for Run %d: %.9f seconds", i+1, concurrencyOverhead.Seconds()))
		err = AppendToLogFile(logFilePath, fmt.Sprintf("Memory Usage for Run %d: %.9f MB", i+1, float64(memoryUsage)/(1024*1024)))
		err = AppendToLogFile(logFilePath, fmt.Sprintf("CPU Utilization for Run %d: %.9f%%", i+1, cpuUsage))
	}

	err = AppendToLogFile(logFilePath, "\nAverage Metrics:")
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Average Execution Time: %.9f seconds", totalExecutionTime.Seconds()/float64(numRuns)))
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Average Concurrency Overhead: %.9f seconds", totalConcurrencyOverhead.Seconds()/float64(numRuns)))
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Average Memory Usage: %.9f MB", float64(totalMemoryUsage)/(float64(numRuns)*1024*1024)))
	err = AppendToLogFile(logFilePath, fmt.Sprintf("Average CPU Utilization: %.9f%%", totalCPUUsage/float64(numRuns)))
}
