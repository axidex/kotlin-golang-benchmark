package main

import (
	"context"
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
				// For mixed operations, execute full CRUD cycle
				if task.Type == MixedOperations {
					executeCRUDCycle(ctx, &totalRequests, &successRequests, &failedRequests, &latencies, &latenciesMutex)
				} else {
					// Single operation mode
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
			}
		}()
	}

	// Start request generator
	startTime := time.Now()

	// For mixed operations, each cycle contains 4 HTTP requests
	// So we need to divide RPS by 4 to get the correct number of cycles
	rps := config.RPS
	if config.BenchmarkType == MixedOperations {
		rps = config.RPS / 4
		if rps == 0 {
			rps = 1 // Minimum 1 cycle per second
		}
	}

	ticker := time.NewTicker(time.Second / time.Duration(rps))
	defer ticker.Stop()

	benchmarkCtx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	// Generate requests at specified RPS
	go func() {
		defer close(requestQueue)
		for {
			select {
			case <-benchmarkCtx.Done():
				return
			case <-ticker.C:
				// For mixed operations, pass MixedOperations type
				// The worker will execute full CRUD cycle
				requestQueue <- RequestTask{Type: config.BenchmarkType}
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
		BenchmarkType:   config.BenchmarkType,
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

// executeCRUDCycle executes full CRUD cycle: CREATE -> GET -> UPDATE -> DELETE
// This counts as ONE request iteration
func executeCRUDCycle(ctx *RequestContext, totalRequests, successRequests, failedRequests *int64, latencies *[]time.Duration, mutex *sync.Mutex) {
	atomic.AddInt64(totalRequests, 1)

	start := time.Now()
	var cycleSuccess = true

	// Step 1: CREATE product and get ID
	productID, err := createProductAndGetID(ctx)
	if err != nil {
		atomic.AddInt64(failedRequests, 1)
		cycleSuccess = false
		// Record latency even for failed cycle
		mutex.Lock()
		*latencies = append(*latencies, time.Since(start))
		mutex.Unlock()
		return
	}

	// Step 2: GET product by ID
	_, _, err = getProductByID(ctx, int(productID))
	if err != nil {
		cycleSuccess = false
	}

	// Step 3: UPDATE product
	_, _, err = updateProduct(ctx, int(productID))
	if err != nil {
		cycleSuccess = false
	}

	// Step 4: DELETE product
	_, _, err = deleteProduct(ctx, int(productID))
	if err != nil {
		cycleSuccess = false
	}

	latency := time.Since(start)

	mutex.Lock()
	*latencies = append(*latencies, latency)
	mutex.Unlock()

	if cycleSuccess {
		atomic.AddInt64(successRequests, 1)
	} else {
		atomic.AddInt64(failedRequests, 1)
	}
}
