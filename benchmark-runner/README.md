# Benchmark Runner

Go-based benchmark tool for load testing Quarkus and Golang applications.

## Features

- Configurable RPS (Requests Per Second)
- Multiple benchmark types (GET, POST, PUT, DELETE)
- Concurrent workers with honest RPS counting
- Detailed latency statistics (min, avg, max, p50, p95, p99)
- Detailed error reporting with grouping by error type
- JSON output for automated processing

## Build

```bash
cd benchmark-runner
go build -o benchmark-runner main.go
```

## Usage

```bash
# Run GET products benchmark
./benchmark-runner \
  -url=http://localhost:8080 \
  -type=get-products \
  -rps=100 \
  -duration=30s \
  -concurrency=10

# Run POST create product benchmark
./benchmark-runner \
  -url=http://localhost:8080 \
  -type=create-product \
  -rps=50 \
  -duration=1m \
  -concurrency=5

# Run UPDATE product benchmark
./benchmark-runner \
  -url=http://localhost:8080 \
  -type=update-product \
  -rps=100 \
  -duration=1m \
  -concurrency=10

# Run mixed operations benchmark (realistic workload)
./benchmark-runner \
  -url=http://localhost:8080 \
  -type=mixed-operations \
  -rps=100 \
  -duration=1m \
  -concurrency=10

# Enable verbose error logging
./benchmark-runner \
  -url=http://localhost:8080 \
  -type=get-products \
  -rps=500 \
  -duration=30s \
  -concurrency=20 \
  -verbose
```

## Command Line Options

- `-url` - Target URL (required)
- `-type` - Benchmark type: `get-products`, `create-product`, `get-product-by-id`, `update-product`, `delete-product`, `mixed-operations` (default: `get-products`)
- `-rps` - Requests per second (default: `100`)
- `-duration` - Duration (e.g., `30s`, `1m`, `5m`) (default: `30s`)
- `-concurrency` - Number of concurrent workers (default: `10`)
- `-verbose` - Enable verbose error logging with response bodies (default: `false`)

## Benchmark Types

All operations use fixed product ID (1) for testing. 4xx errors (404, 409) are considered "successful" since they reach the API and database - we only care about 5xx errors.

1. **get-products** - GET request to `/api/products` (fast, ~10-50ms latency)
2. **create-product** - POST request to `/api/products` with product JSON (medium, ~50-100ms latency)
3. **get-product-by-id** - GET request to `/api/products/1` (fast, ~10-30ms latency)
4. **update-product** - PUT request to `/api/products/1` with updated product JSON (medium, ~50-100ms latency)
5. **delete-product** - DELETE request to `/api/products/1` (fast, ~10-30ms latency)
6. **mixed-operations** - Randomly selects from CREATE/GET by ID/UPDATE/DELETE operations for each request (realistic CRUD workload with fixed product ID)

## Concurrency Recommendations

The concurrency parameter depends on your target RPS and expected latency:

**Formula:** `concurrency ≈ (RPS × average_latency_seconds)`

Examples:
- **get-products** (20ms latency):
  - RPS=100 → concurrency=2-5
  - RPS=1000 → concurrency=20-50
- **create-product** (50ms latency):
  - RPS=100 → concurrency=5-10
  - RPS=1000 → concurrency=50-100
- **update-product** (50ms latency):
  - RPS=100 → concurrency=5-10
  - RPS=1000 → concurrency=50-100

**Note:** If actual RPS is much lower than target RPS, increase concurrency. Workers are blocking on I/O, so higher concurrency is needed for long-running operations.

## Output

The benchmark outputs:
- Total/Success/Failed request counts
- Actual RPS achieved
- Latency statistics (min, avg, max, p50, p95, p99)
- Top 10 most common errors (grouped by type)
- JSON formatted results for automation

Example output:
```
====================================
Benchmark Results
====================================
Total Requests:   6000
Success:          6000 (100.00%)
Failed:           0 (0.00%)
Duration:         1m0s
Actual RPS:       100.00 req/s

Latency:
  Min:            2.5ms
  Avg:            15.3ms
  Max:            125.8ms
  P50:            12.1ms
  P95:            35.2ms
  P99:            58.7ms
====================================

JSON Results:
{
  "total_requests": 6000,
  "success_requests": 6000,
  "failed_requests": 0,
  "duration_seconds": 60.0,
  "rps": 100.0,
  "latency": {
    "min": "2.5ms",
    "avg": "15.3ms",
    "max": "125.8ms",
    "p50": "12.1ms",
    "p95": "35.2ms",
    "p99": "58.7ms"
  }
}
```

## Monitoring Results

After running benchmarks, check:
- Grafana dashboards for detailed metrics
- Prometheus for raw metrics data
- Application logs for errors
