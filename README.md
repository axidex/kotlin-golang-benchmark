# Kotlin Quarkus vs Golang Gin Benchmark

Project for comparing performance of CRUD applications on Quarkus (Kotlin) and Gin (Golang) using PostgreSQL, Prometheus and Grafana.

## Architecture

```
┌─────────────────┐     ┌─────────────────┐
│  Quarkus App    │     │   Golang App    │
│  (Kotlin)       │     │   (Gin)         │
│  Port: 8080     │     │   Port: 8081    │
└────────┬────────┘     └────────┬────────┘
         │                       │
         ├───────────┬───────────┤
         │           │           │
    ┌────▼────┐ ┌───▼──────┐ ┌──▼────────┐
    │PostgreSQL│ │Prometheus│ │  Grafana  │
    │Port: 5432│ │Port: 9090│ │Port: 3000 │
    └──────────┘ └──────────┘ └───────────┘
```

## Quick start

### 1. Start all services

```bash
docker-compose up --build
```

This will start:
- **PostgreSQL** (port 5432) - shared database for both applications
- **Quarkus App** (port 8080) - Kotlin CRUD API
- **Golang App** (port 8081) - Go CRUD API
- **Prometheus** (port 9090) - metrics collection
- **Grafana** (port 3000) - metrics visualization

### 2. Access to services

| Service | URL | Credentials |
|--------|-----|-------------|
| Quarkus API | http://localhost:8080/api/products | - |
| Golang API | http://localhost:8081/api/products | - |
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | - |
| Quarkus Metrics | http://localhost:8080/q/metrics | - |
| Golang Metrics | http://localhost:8081/metrics | - |

### 3. Run benchmark

```bash
./benchmark.sh
```

## API Endpoints

Both applications provide identical REST API:

- `GET /api/products` - get all products
- `GET /api/products/:id` - get product by ID
- `POST /api/products` - create new product
- `PUT /api/products/:id` - update product
- `DELETE /api/products/:id` - delete product

### Request examples

```bash
# Create product in Quarkus
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","description":"Gaming laptop","price":1500.00,"quantity":10}'

# Create product in Golang
curl -X POST http://localhost:8081/api/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Mouse","description":"Gaming mouse","price":50.00,"quantity":100}'

# Get all products
curl http://localhost:8080/api/products
curl http://localhost:8081/api/products
```

## Metrics in Grafana

After starting, open Grafana at http://localhost:3000 (admin/admin).

Dashboard "Kotlin Quarkus vs Golang Gin Benchmark" contains:

### Performance panels
- **Requests per Second** - number of requests per second for each application
- **Average Response Time** - average response time in milliseconds
- **HTTP Status Codes** - distribution of HTTP response codes

### Resource usage panels
- **Memory Usage** - memory consumption (Heap for JVM, Alloc for Go)
- **Total Requests** - total number of processed requests
- **Active Goroutines** - number of active goroutines (Go only)

## Project structure

```
.
├── kotlin-quarkus/              # Quarkus application
│   ├── src/main/kotlin/
│   │   └── dev/sourcecraft/dolgintsev/
│   │       ├── entity/Product.kt
│   │       └── resource/ProductResource.kt
│   ├── src/main/resources/
│   │   └── application.properties
│   ├── build.gradle.kts
│   └── Dockerfile
│
├── golang-gin/                  # Golang application
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── models/product.go
│   │   ├── handlers/product_handler.go
│   │   ├── database/database.go
│   │   └── middleware/metrics.go
│   ├── go.mod
│   └── Dockerfile
│
├── grafana/                     # Grafana configuration
│   └── provisioning/
│       ├── datasources/prometheus.yml
│       └── dashboards/
│           ├── dashboard.yml
│           └── benchmark-dashboard.json
│
├── prometheus.yml               # Prometheus configuration
├── docker-compose.yml           # Docker compose for all services
└── benchmark.sh                 # Benchmark script
```

## Technologies

### Quarkus (Kotlin)
- Quarkus 3.17.6
- Hibernate ORM Panache (Kotlin)
- PostgreSQL JDBC Driver
- Micrometer + Prometheus
- SmallRye Health

### Golang (Gin)
- Gin Web Framework
- GORM (PostgreSQL)
- Prometheus client
- Custom metrics middleware

## Stop and cleanup

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (clean database data)
docker-compose down -v
```

## Benchmarking Tips

1. **Warmup**: Run several requests before starting benchmark to warm up JVM
2. **Load**: Vary `-n` (requests) and `-c` (concurrency) parameters in benchmark.sh
3. **Monitoring**: Watch metrics in Grafana in real-time during benchmark
4. **Resources**: Make sure Docker has enough resources (CPU, RAM)

## Results

After running benchmark, compare:
- Throughput (requests per second)
- Latency (response time)
- Memory usage
- Stability under load