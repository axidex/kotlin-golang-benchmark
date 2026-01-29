package main

import (
	"net/http"
	"time"
)

type BenchmarkType string

const (
	GetProducts      BenchmarkType = "get-products"
	CreateProduct    BenchmarkType = "create-product"
	GetProductByID   BenchmarkType = "get-product-by-id"
	UpdateProduct    BenchmarkType = "update-product"
	DeleteProduct    BenchmarkType = "delete-product"
	MixedOperations  BenchmarkType = "mixed-operations"
)

type Config struct {
	URL           string
	RPS           int
	Duration      time.Duration
	BenchmarkType BenchmarkType
	Concurrency   int
	Verbose       bool
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
	Errors          *ErrorStats
}

type Product struct {
	ID          int64   `json:"id,omitempty"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

type RequestContext struct {
	Client     *http.Client
	Config     Config
	ErrorStats *ErrorStats
}

type RequestTask struct {
	Type BenchmarkType
}