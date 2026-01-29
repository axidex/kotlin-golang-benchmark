package main

import (
	"context"
	"math/rand"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// runBenchmark runs the benchmark with honest RPS counting
// It creates a queue of requests and workers that process them
func runBenchmark(config Config) *Result {
	var (
		totalRequests   int64
		successRequests int64
		failedRequests  int64
		latencies       []time.Duration
		latenciesMutex  sync.Mutex
	)

	errorStats := NewErrorStats()

	// Create context for request execution
	ctx := &RequestContext{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Config:     config,
		ErrorStats: errorStats,
	}

	// Create request queue channel
	requestQueue := make(chan RequestTask, config.Concurrency*2)

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range requestQueue {
				atomic.AddInt64(&totalRequests, 1)

				latency, err := executeRequest(ctx, task)

				latenciesMutex.Lock()
				latencies = append(latencies, latency)
				latenciesMutex.Unlock()

				if err != nil {
					atomic.AddInt64(&failedRequests, 1)
				} else {
					atomic.AddInt64(&successRequests, 1)
				}
			}
		}()
	}

	// Start request generator
	startTime := time.Now()
	ticker := time.NewTicker(time.Second / time.Duration(config.RPS))
	defer ticker.Stop()

	benchmarkCtx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	// All available operations for mixed mode (excluding GetProducts)
	allOperations := []BenchmarkType{
		CreateProduct,
		GetProductByID,
		UpdateProduct,
		DeleteProduct,
	}

	// Generate requests at specified RPS
	go func() {
		defer close(requestQueue)
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for {
			select {
			case <-benchmarkCtx.Done():
				return
			case <-ticker.C:
				taskType := config.BenchmarkType
				// If mixed operations, randomly select operation type
				if config.BenchmarkType == MixedOperations {
					taskType = allOperations[rng.Intn(len(allOperations))]
				}
				requestQueue <- RequestTask{Type: taskType}
			}
		}
	}()

	// Wait for context to expire
	<-benchmarkCtx.Done()

	// Wait for all workers to finish processing remaining requests
	wg.Wait()

	duration := time.Since(startTime)

	// Calculate statistics
	result := &Result{
		TotalRequests:   totalRequests,
		SuccessRequests: successRequests,
		FailedRequests:  failedRequests,
		TotalDuration:   duration,
		Latencies:       latencies,
		Errors:          errorStats,
	}

	calculateLatencyStats(result)

	return result
}

// calculateLatencyStats calculates latency statistics
func calculateLatencyStats(result *Result) {
	if len(result.Latencies) == 0 {
		return
	}

	// Sort latencies for percentile calculation
	sort.Slice(result.Latencies, func(i, j int) bool {
		return result.Latencies[i] < result.Latencies[j]
	})

	result.MinLatency = result.Latencies[0]
	result.MaxLatency = result.Latencies[len(result.Latencies)-1]

	// Calculate average
	var sum time.Duration
	for _, lat := range result.Latencies {
		sum += lat
	}
	result.AvgLatency = sum / time.Duration(len(result.Latencies))

	// Calculate percentiles
	result.P50Latency = percentile(result.Latencies, 0.50)
	result.P95Latency = percentile(result.Latencies, 0.95)
	result.P99Latency = percentile(result.Latencies, 0.99)
}

// percentile calculates the percentile value from sorted latencies
func percentile(sortedLatencies []time.Duration, p float64) time.Duration {
	if len(sortedLatencies) == 0 {
		return 0
	}
	index := int(float64(len(sortedLatencies)) * p)
	if index >= len(sortedLatencies) {
		index = len(sortedLatencies) - 1
	}
	return sortedLatencies[index]
}
