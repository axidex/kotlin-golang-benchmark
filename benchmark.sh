#!/bin/bash

# Benchmark script for comparing Quarkus and Gin performance

QUARKUS_URL="http://155.212.170.172:30080"
GOLANG_URL="http://155.212.170.172:30081"
REQUESTS=10000
CONCURRENCY=100

echo "======================================"
echo "Benchmark: Quarkus vs Golang Gin"
echo "======================================"
echo ""

# Check if ab (Apache Bench) is installed
if ! command -v ab &> /dev/null; then
    echo "Error: Apache Bench (ab) is not installed"
    echo "Install with: brew install apache2 (macOS) or apt-get install apache2-utils (Linux)"
    exit 1
fi

# Wait for services to be ready
echo "Waiting for services to start..."

# Function to wait for service to be ready
wait_for_service() {
    local url=$1
    local name=$2
    local max_attempts=30
    local attempt=1

    echo "Waiting for $name to be ready..."
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            echo "$name is ready!"
            return 0
        fi
        echo "Attempt $attempt/$max_attempts: $name not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done

    echo "Error: $name did not become ready in time"
    return 1
}

# Wait for both services
wait_for_service "${QUARKUS_URL}/q/health" "Quarkus" || exit 1
wait_for_service "${GOLANG_URL}/health" "Golang" || exit 1

echo ""
echo "Both services are ready!"

echo ""
echo "======================================"
echo "1. GET All Products Benchmark"
echo "======================================"

echo ""
echo "Testing Quarkus (Kotlin)..."
ab -n ${REQUESTS} -c ${CONCURRENCY} "${QUARKUS_URL}/api/products"

echo ""
echo "Testing Golang (Gin)..."
ab -n ${REQUESTS} -c ${CONCURRENCY} "${GOLANG_URL}/api/products"

echo ""
echo "======================================"
echo "2. POST Create Product Benchmark"
echo "======================================"

PRODUCT_JSON='{"name":"Benchmark Product","description":"Test product for benchmark","price":99.99,"quantity":100}'

echo ""
echo "Testing Quarkus (Kotlin)..."
ab -n 1000 -c 10 -p <(echo ${PRODUCT_JSON}) -T "application/json" "${QUARKUS_URL}/api/products"

echo ""
echo "Testing Golang (Gin)..."
ab -n 1000 -c 10 -p <(echo ${PRODUCT_JSON}) -T "application/json" "${GOLANG_URL}/api/products"

echo ""
echo "======================================"
echo "3. Metrics Comparison"
echo "======================================"

echo ""
echo "Quarkus Metrics:"
curl -s ${QUARKUS_URL}/q/metrics | grep http_server_requests

echo ""
echo "Golang Metrics:"
curl -s ${GOLANG_URL}/metrics | grep http_requests_total

echo ""
echo "======================================"
echo "Benchmark Complete!"
echo "======================================"
echo ""
echo "View detailed metrics:"
echo "  - Grafana Dashboard: http://localhost:3000 (admin/admin)"
echo "  - Prometheus: http://localhost:9090"
echo "  - Quarkus Metrics: ${QUARKUS_URL}/q/metrics"
echo "  - Golang Metrics: ${GOLANG_URL}/metrics"
echo ""
