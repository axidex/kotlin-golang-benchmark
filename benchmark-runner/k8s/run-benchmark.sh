#!/bin/bash

# Script to run benchmark as Kubernetes Job
# Usage: ./run-benchmark.sh [options]

set -e

# Default values
TARGET_APP="quarkus"
BENCHMARK_TYPE="get-products"
RPS="100"
DURATION="1m"
CONCURRENCY="10"
KUBECONFIG="${KUBECONFIG:-~/.kube/yacloud-k3s.yaml}"
NAMESPACE="benchmark"

# Help function
show_help() {
    cat << EOF
Usage: ${0##*/} [OPTIONS]

Run benchmark as Kubernetes Job

OPTIONS:
    -a, --app APP           Target app: quarkus, quarkus-native, golang (default: quarkus)
    -t, --type TYPE         Benchmark type: get-products, create-product, get-product-by-id, mixed-crud (default: get-products)
    -r, --rps RPS           Requests per second (default: 100)
    -d, --duration DURATION Duration, e.g., 30s, 1m, 5m (default: 1m)
    -c, --concurrency NUM   Concurrent workers (default: 10)
    -k, --kubeconfig PATH   Path to kubeconfig (default: ~/.kube/yacloud-k3s.yaml)
    -n, --namespace NS      Kubernetes namespace (default: benchmark)
    -h, --help              Show this help message

EXAMPLES:
    # Run GET products benchmark on Quarkus with 100 RPS for 1 minute
    ${0##*/} -a quarkus -t get-products -r 100 -d 1m

    # Run POST benchmark on Golang with 200 RPS for 30 seconds
    ${0##*/} -a golang -t create-product -r 200 -d 30s -c 20

    # Full CRUD cycle benchmark (create, get, update, delete)
    ${0##*/} -a quarkus -t mixed-crud -r 50 -d 1m -c 5

    # High load test on native Quarkus
    ${0##*/} -a quarkus-native -r 1000 -d 5m -c 50

EOF
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -a|--app)
            TARGET_APP="$2"
            shift 2
            ;;
        -t|--type)
            BENCHMARK_TYPE="$2"
            shift 2
            ;;
        -r|--rps)
            RPS="$2"
            shift 2
            ;;
        -d|--duration)
            DURATION="$2"
            shift 2
            ;;
        -c|--concurrency)
            CONCURRENCY="$2"
            shift 2
            ;;
        -k|--kubeconfig)
            KUBECONFIG="$2"
            shift 2
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Determine target URL based on app
case $TARGET_APP in
    quarkus)
        TARGET_URL="http://benchmark-quarkus:8080"
        ;;
    quarkus-native)
        TARGET_URL="http://benchmark-quarkus-native:8080"
        ;;
    golang)
        TARGET_URL="http://benchmark-golang:8080"
        ;;
    *)
        echo "Error: Unknown app '$TARGET_APP'. Use: quarkus, quarkus-native, or golang"
        exit 1
        ;;
esac

# Generate unique benchmark name
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BENCHMARK_NAME="${TARGET_APP}-${BENCHMARK_TYPE}-${TIMESTAMP}"

echo "======================================"
echo "Starting Benchmark Job"
echo "======================================"
echo "App:         ${TARGET_APP}"
echo "Type:        ${BENCHMARK_TYPE}"
echo "RPS:         ${RPS}"
echo "Duration:    ${DURATION}"
echo "Concurrency: ${CONCURRENCY}"
echo "Job Name:    ${BENCHMARK_NAME}"
echo "======================================"
echo ""

# Generate job manifest from template
TEMP_MANIFEST=$(mktemp)
export BENCHMARK_NAME TARGET_APP BENCHMARK_TYPE TARGET_URL RPS DURATION CONCURRENCY
envsubst < "$(dirname "$0")/job-template.yaml" > "${TEMP_MANIFEST}"

echo "Creating Kubernetes Job..."
kubectl apply -f "${TEMP_MANIFEST}" --kubeconfig="${KUBECONFIG}"

# Clean up temp file
rm -f "${TEMP_MANIFEST}"

echo ""
echo "Job created successfully!"
echo ""
echo "Monitor job status:"
echo "  kubectl get job ${BENCHMARK_NAME} -n ${NAMESPACE} --kubeconfig=${KUBECONFIG}"
echo ""
echo "View logs:"
echo "  kubectl logs -f job/${BENCHMARK_NAME} -n ${NAMESPACE} --kubeconfig=${KUBECONFIG}"
echo ""
echo "Delete job when done:"
echo "  kubectl delete job ${BENCHMARK_NAME} -n ${NAMESPACE} --kubeconfig=${KUBECONFIG}"
echo ""

# Optionally follow logs
read -p "Follow logs now? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Waiting for pod to start..."
    kubectl wait --for=condition=Ready pod -l job-name="${BENCHMARK_NAME}" -n "${NAMESPACE}" --kubeconfig="${KUBECONFIG}" --timeout=60s || true
    echo ""
    kubectl logs -f "job/${BENCHMARK_NAME}" -n "${NAMESPACE}" --kubeconfig="${KUBECONFIG}"
fi