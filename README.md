# Java-Go-Concurrency-Benchmarks

## Overview
This repository contains the implementation and benchmarking of concurrency models in **Java** and **Go**. Specifically, it compares **Java’s thread-based concurrency** using **ExecutorService** and **BlockingQueue** with **Go’s goroutines**. The goal is to evaluate the performance of these two concurrency models in handling image classification tasks using **Convolutional Neural Networks (CNNs)**, with benchmarks conducted on the **CIFAR-10** and **Tiny ImageNet** datasets.

## Structure
The repository is organized into the following sections:

### 1. `java/`
Contains the Java implementation of the **Producer-Consumer model** for concurrent image classification, using **DeepLearning4j (DL4J)** for building and training CNNs.
- **Key components**:
  - `ImageTask.java`: Defines the task to be run by each thread.
  - `BlockingQueue.java`: Manages task queueing.
  - `ExecutorService.java`: Manages the thread pool and task execution.
  
### 2. `go/`
Contains the Go implementation of the **Producer-Consumer model** using **Gorgonia** for CNN training and inference, utilizing **goroutines** and **channels** for concurrency.
- **Key components**:
  - `cnn.go`: Implements the CNN using Gorgonia.
  - `concurrency.go`: Manages goroutines and task distribution.
  
### 3. `datasets/`
Contains the datasets used for benchmarking image classification tasks.
- Links to datasets:
  - [CIFAR-10 Dataset](https://www.cs.toronto.edu/~kriz/cifar.html)
  - [Tiny ImageNet Dataset](https://www.kaggle.com/datasets/akash2sharma/tiny-imagenet)

### 4. `benchmarks/`
Scripts and result files for benchmarking Java’s thread pool and Go’s goroutines across various metrics:
- **Execution Time**
- **Memory Usage**
- **CPU Utilization**
- **Concurrency Overhead**

