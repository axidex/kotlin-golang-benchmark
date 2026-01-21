package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type BenchmarkType string

const (
	GetProducts    BenchmarkType = "get-products"
	CreateProduct  BenchmarkType = "create-product"
	GetProductByID BenchmarkType = "get-product-by-id"
	MixedCRUD      BenchmarkType = "mixed-crud"
)

type Config struct {
	URL           string
	RPS           int
	Duration      time.Duration
	BenchmarkType BenchmarkType
	Concurrency   int
}

type Result struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64
	TotalDuration   time.Duration
	MinLatency      time.Duration
	MaxLatency      time.Duration
	AvgLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	Latencies       []time.Duration
}

type Product struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

func main() {
	var config Config

	flag.StringVar(&config.URL, "url", "", "Target URL (required)")
	flag.IntVar(&config.RPS, "rps", 100, "Requests per second")
	durationStr := flag.String("duration", "30s", "Benchmark duration (e.g., 30s, 1m, 5m)")
	benchType := flag.String("type", string(GetProducts), "Benchmark type: get-products, create-product, get-product-by-id, mixed-crud")
	flag.IntVar(&config.Concurrency, "concurrency", 10, "Number of concurrent workers")
	flag.Parse()

	if config.URL == "" {
		log.Fatal("URL is required. Use -url flag")
	}

	duration, err := time.ParseDuration(*durationStr)
	if err != nil {
		log.Fatalf("Invalid duration: %v", err)
	}
	config.Duration = duration
	config.BenchmarkType = BenchmarkType(*benchType)

	log.Printf("Starting benchmark:")
	log.Printf("  URL: %s", config.URL)
	log.Printf("  Type: %s", config.BenchmarkType)
	log.Printf("  RPS: %d", config.RPS)
	log.Printf("  Duration: %s", config.Duration)
	log.Printf("  Concurrency: %d", config.Concurrency)
	log.Printf("")

	result := runBenchmark(config)
	printResults(result)
}

func runBenchmark(config Config) *Result {
	var (
		totalRequests   int64
		successRequests int64
		failedRequests  int64
		latencies       []time.Duration
		latenciesMutex  sync.Mutex
	)

	// Calculate requests per worker
	ticker := time.NewTicker(time.Second / time.Duration(config.RPS))
	defer ticker.Stop()

	// Create worker pool
	var wg sync.WaitGroup
	requestChan := make(chan struct{}, config.Concurrency)

	// Benchmark context
	ctx, cancel := context.WithTimeout(context.Background(), config.Duration)
	defer cancel()

	startTime := time.Now()

	// Start workers
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client := &http.Client{
				Timeout: 10 * time.Second,
			}

			for range requestChan {
				reqStart := time.Now()
				success := executeRequest(client, config)
				latency := time.Since(reqStart)

				atomic.AddInt64(&totalRequests, 1)
				if success {
					atomic.AddInt64(&successRequests, 1)
				} else {
					atomic.AddInt64(&failedRequests, 1)
				}

				latenciesMutex.Lock()
				latencies = append(latencies, latency)
				latenciesMutex.Unlock()
			}
		}()
	}

	// Send requests at specified RPS
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(requestChan)
				return
			case <-ticker.C:
				select {
				case requestChan <- struct{}{}:
				default:
					// Workers are busy, skip this tick
				}
			}
		}
	}()

	// Wait for duration
	time.Sleep(config.Duration)
	cancel()
	wg.Wait()

	totalDuration := time.Since(startTime)

	return calculateResults(totalRequests, successRequests, failedRequests, totalDuration, latencies)
}

func executeRequest(client *http.Client, config Config) bool {
	switch config.BenchmarkType {
	case GetProducts:
		return executeGetProducts(client, config.URL)
	case CreateProduct:
		return executeCreateProduct(client, config.URL)
	case GetProductByID:
		return executeGetProductByID(client, config.URL, 1)
	case MixedCRUD:
		return executeMixedCRUD(client, config.URL)
	default:
		return false
	}
}

func executeGetProducts(client *http.Client, baseURL string) bool {
	req, err := http.NewRequest("GET", baseURL+"/api/products", nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func executeCreateProduct(client *http.Client, baseURL string) bool {
	product := Product{
		Name:        "Benchmark Product",
		Description: "Test product for benchmark",
		Price:       99.99,
		Quantity:    100,
	}
	body, _ := json.Marshal(product)

	req, err := http.NewRequest("POST", baseURL+"/api/products", bytes.NewBuffer(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func executeGetProductByID(client *http.Client, baseURL string, id int64) bool {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/products/%d", baseURL, id), nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func executeMixedCRUD(client *http.Client, baseURL string) bool {
	// 1. Create product
	product := Product{
		Name:        "CRUD Test Product",
		Description: "Product for mixed CRUD benchmark",
		Price:       49.99,
		Quantity:    50,
	}
	body, _ := json.Marshal(product)

	req, err := http.NewRequest("POST", baseURL+"/api/products", bytes.NewBuffer(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return false
	}

	// Parse created product to get ID
	var createdProduct struct {
		ID          int64   `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Quantity    int     `json:"quantity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createdProduct); err != nil {
		resp.Body.Close()
		return false
	}
	resp.Body.Close()

	productID := createdProduct.ID

	// 2. Get created product
	if !executeGetProductByID(client, baseURL, productID) {
		return false
	}

	// 3. Update product
	updatedProduct := Product{
		Name:        "Updated CRUD Product",
		Description: "Updated description",
		Price:       59.99,
		Quantity:    75,
	}
	body, _ = json.Marshal(updatedProduct)

	req, err = http.NewRequest("PUT", fmt.Sprintf("%s/api/products/%d", baseURL, productID), bytes.NewBuffer(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		return false
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false
	}

	// 4. Delete product
	req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/api/products/%d", baseURL, productID), nil)
	if err != nil {
		return false
	}

	resp, err = client.Do(req)
	if err != nil {
		return false
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func calculateResults(total, success, failed int64, duration time.Duration, latencies []time.Duration) *Result {
	if len(latencies) == 0 {
		return &Result{
			TotalRequests:   total,
			SuccessRequests: success,
			FailedRequests:  failed,
			TotalDuration:   duration,
		}
	}

	// Sort latencies for percentile calculation
	sortLatencies(latencies)

	var sum time.Duration
	minLatency := latencies[0]
	maxLatency := latencies[0]

	for _, l := range latencies {
		sum += l
		if l < minLatency {
			minLatency = l
		}
		if l > maxLatency {
			maxLatency = l
		}
	}

	avg := sum / time.Duration(len(latencies))
	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]

	return &Result{
		TotalRequests:   total,
		SuccessRequests: success,
		FailedRequests:  failed,
		TotalDuration:   duration,
		MinLatency:      minLatency,
		MaxLatency:      maxLatency,
		AvgLatency:      avg,
		P50Latency:      p50,
		P95Latency:      p95,
		P99Latency:      p99,
		Latencies:       latencies,
	}
}

func sortLatencies(latencies []time.Duration) {
	// Simple bubble sort (good enough for benchmark results)
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}
}

func printResults(r *Result) {
	fmt.Println("")
	fmt.Println("====================================")
	fmt.Println("Benchmark Results")
	fmt.Println("====================================")
	fmt.Printf("Total Requests:   %d\n", r.TotalRequests)
	fmt.Printf("Success:          %d (%.2f%%)\n", r.SuccessRequests, float64(r.SuccessRequests)/float64(r.TotalRequests)*100)
	fmt.Printf("Failed:           %d (%.2f%%)\n", r.FailedRequests, float64(r.FailedRequests)/float64(r.TotalRequests)*100)
	fmt.Printf("Duration:         %s\n", r.TotalDuration)
	fmt.Printf("Actual RPS:       %.2f req/s\n", float64(r.TotalRequests)/r.TotalDuration.Seconds())
	fmt.Println("")
	fmt.Println("Latency:")
	fmt.Printf("  Min:            %s\n", r.MinLatency)
	fmt.Printf("  Avg:            %s\n", r.AvgLatency)
	fmt.Printf("  Max:            %s\n", r.MaxLatency)
	fmt.Printf("  P50:            %s\n", r.P50Latency)
	fmt.Printf("  P95:            %s\n", r.P95Latency)
	fmt.Printf("  P99:            %s\n", r.P99Latency)
	fmt.Println("====================================")

	// Export results as JSON for automated processing
	jsonResult, _ := json.MarshalIndent(map[string]interface{}{
		"total_requests":   r.TotalRequests,
		"success_requests": r.SuccessRequests,
		"failed_requests":  r.FailedRequests,
		"duration_seconds": r.TotalDuration.Seconds(),
		"rps":              float64(r.TotalRequests) / r.TotalDuration.Seconds(),
		"latency": map[string]string{
			"min": r.MinLatency.String(),
			"avg": r.AvgLatency.String(),
			"max": r.MaxLatency.String(),
			"p50": r.P50Latency.String(),
			"p95": r.P95Latency.String(),
			"p99": r.P99Latency.String(),
		},
	}, "", "  ")
	fmt.Println("")
	fmt.Println("JSON Results:")
	fmt.Println(string(jsonResult))
}
