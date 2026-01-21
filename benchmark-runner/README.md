# Benchmark Runner

Go-based benchmark tool for load testing Quarkus and Golang applications in Kubernetes.

## Features

- Configurable RPS (Requests Per Second)
- Multiple benchmark types (GET, POST)
- Concurrent workers
- Detailed latency statistics (min, avg, max, p50, p95, p99)
- JSON output for automated processing
- Runs as Kubernetes Job

## Build and Push Docker Image

```bash
cd benchmark-runner
docker build -t axidex/benchmark-runner:latest .
docker push axidex/benchmark-runner:latest
```

## Run Locally

```bash
# Build
go build -o benchmark-runner main.go

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
```

## Run in Kubernetes

### Quick Start

```bash
cd k8s

# Make script executable
chmod +x run-benchmark.sh

# Run benchmark on Quarkus JVM
./run-benchmark.sh -a quarkus -t get-products -r 100 -d 1m

# Run benchmark on Golang
./run-benchmark.sh -a golang -t create-product -r 200 -d 30s -c 20

# High load test on Quarkus Native
./run-benchmark.sh -a quarkus-native -r 1000 -d 5m -c 50
```

### Options

- `-a, --app` - Target app: `quarkus`, `quarkus-native`, `golang`
- `-t, --type` - Benchmark type: `get-products`, `create-product`, `get-product-by-id`
- `-r, --rps` - Requests per second
- `-d, --duration` - Duration (e.g., `30s`, `1m`, `5m`)
- `-c, --concurrency` - Number of concurrent workers
- `-k, --kubeconfig` - Path to kubeconfig file
- `-n, --namespace` - Kubernetes namespace

### Manual Job Creation

```bash
# Edit job-template.yaml with your parameters
export TARGET_URL="http://benchmark-quarkus:8080"
export BENCHMARK_TYPE="get-products"
export RPS="100"
export DURATION="1m"
export CONCURRENCY="10"
export BENCHMARK_NAME="test-$(date +%s)"
export TARGET_APP="quarkus"

envsubst < job-template.yaml | kubectl apply -f - --kubeconfig=~/.kube/yacloud-k3s.yaml

# View logs
kubectl logs -f job/benchmark-test-$(date +%s) -n benchmark --kubeconfig=~/.kube/yacloud-k3s.yaml
```

## Benchmark Types

1. **get-products** - GET request to `/api/products` (fast, ~10-50ms latency)
2. **create-product** - POST request to `/api/products` with product JSON (medium, ~50-100ms latency)
3. **get-product-by-id** - GET request to `/api/products/1` (fast, ~10-30ms latency)
4. **mixed-crud** - Full CRUD cycle: CREATE → GET → UPDATE → DELETE (slow, ~300-500ms latency, requires high concurrency)

### Concurrency Recommendations

The concurrency parameter depends on your target RPS and expected latency:

**Formula:** `concurrency ≈ (RPS × average_latency_seconds)`

Examples:
- **get-products** (20ms latency):
  - RPS=100 → concurrency=2-5
  - RPS=1000 → concurrency=20-50
- **create-product** (50ms latency):
  - RPS=100 → concurrency=5-10
  - RPS=1000 → concurrency=50-100
- **mixed-crud** (300ms latency):
  - RPS=50 → concurrency=15-20
  - RPS=100 → concurrency=30-50
  - RPS=500 → concurrency=150-200
  - RPS=1000 → concurrency=300-400

**Note:** If actual RPS is much lower than target RPS, increase concurrency. Workers are blocking on I/O, so higher concurrency is needed for long-running operations.

## Output

The benchmark outputs:
- Total/Success/Failed request counts
- Actual RPS achieved
- Latency statistics (min, avg, max, p50, p95, p99)
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
```

## Monitoring Results

After running benchmarks, check:
- Grafana dashboards for detailed metrics
- Prometheus for raw metrics data
- Application logs for errors

## Cleanup

Jobs are automatically cleaned up after 1 hour (ttlSecondsAfterFinished: 3600).

Manual cleanup:
```bash
# Delete specific job
kubectl delete job benchmark-quarkus-get-products-20260120-123456 -n benchmark

# Delete all benchmark jobs
kubectl delete jobs -l app=benchmark-runner -n benchmark
```