package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	var config Config

	flag.StringVar(&config.URL, "url", "", "Target URL (required)")
	flag.IntVar(&config.RPS, "rps", 100, "Requests per second")
	durationStr := flag.String("duration", "30s", "Benchmark duration (e.g., 30s, 1m, 5m)")
	benchType := flag.String("type", string(GetProducts), "Benchmark type: get-products, create-product, get-product-by-id, update-product, delete-product, mixed-operations")
	flag.IntVar(&config.Concurrency, "concurrency", 10, "Number of concurrent workers")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose error logging (show response bodies)")
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
	printResults(result, config.Verbose)
}