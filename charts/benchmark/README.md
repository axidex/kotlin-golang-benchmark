# Benchmark Helm Chart

Helm chart for deploying Kotlin Quarkus and Golang Gin applications in Kubernetes (k3s).

## Installation

### 1. Install chart

```bash
# Basic installation
helm install benchmark . -n benchmark --create-namespace

# With custom values
helm install benchmark . -n benchmark --create-namespace \
  --set ingress.hosts[0].host=quarkus.yourdomain.com \
  --set ingress.hosts[1].host=golang.yourdomain.com
```

### 4. Update existing release

```bash
helm upgrade benchmark . -n benchmark
```

### 5. Remove release

```bash
helm uninstall benchmark -n benchmark
```

## Configuration

### Main parameters

| Parameter | Description | Default value |
|----------|----------|----------------------|
| `quarkus.enabled` | Enable Quarkus application | `true` |
| `quarkus.image.repository` | Quarkus Docker image | `axidex/benchmark-kotlin-quarkus` |
| `quarkus.image.tag` | Image tag | `latest` |
| `golang.enabled` | Enable Golang application | `true` |
| `golang.image.repository` | Golang Docker image | `axidex/benchmark-golang-gin` |
| `golang.image.tag` | Image tag | `latest` |
| `postgresql.enabled` | Enable PostgreSQL | `true` |
| `postgresql.auth.database` | Database name | `benchmark` |
| `postgresql.auth.username` | Database user | `postgres` |
| `postgresql.auth.password` | Database password | `postgres` |

### Ingress

By default, Traefik (built into k3s) is used.

Change hosts in `values.yaml`:

```yaml
ingress:
  hosts:
    - host: quarkus.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
          backend: quarkus
    - host: golang.yourdomain.com
      paths:
        - path: /
          pathType: Prefix
          backend: golang
```

## Usage examples

### Installation without TLS

```bash
helm install benchmark . -n benchmark --create-namespace \
  --set ingress.tls=null
```

### Using LoadBalancer instead of Ingress

```bash
helm install benchmark . -n benchmark --create-namespace \
  --set ingress.enabled=false \
  --set quarkus.service.type=LoadBalancer \
  --set golang.service.type=LoadBalancer
```

### Increasing resources

```bash
helm install benchmark . -n benchmark --create-namespace \
  --set quarkus.resources.limits.memory=1Gi \
  --set quarkus.resources.limits.cpu=2000m
```

## Status check

```bash
# Check pods
kubectl get pods -n benchmark

# Check services
kubectl get svc -n benchmark

# Check ingress
kubectl get ingress -n benchmark

# Quarkus logs
kubectl logs -n benchmark -l app=quarkus -f

# Golang logs
kubectl logs -n benchmark -l app=golang -f
```

## Endpoints

After installation, applications will be available at:

- Quarkus: `http://quarkus.example.com/api/products`
- Golang: `http://golang.example.com/api/products`
- Quarkus Health: `http://quarkus.example.com/q/health`
- Golang Health: `http://golang.example.com/health`
- Quarkus Metrics: `http://quarkus.example.com/q/metrics`
- Golang Metrics: `http://golang.example.com/metrics`

## Troubleshooting

### Check database connection

```bash
kubectl exec -it -n benchmark deployment/benchmark-quarkus -- env | grep DATABASE
kubectl exec -it -n benchmark deployment/benchmark-golang -- env | grep DATABASE
```

### Check PostgreSQL logs

```bash
kubectl logs -n benchmark -l app.kubernetes.io/name=postgresql -f
```

### Port-forward for local access

```bash
# Quarkus
kubectl port-forward -n benchmark svc/benchmark-quarkus 8080:8080

# Golang
kubectl port-forward -n benchmark svc/benchmark-golang 8081:8080

# PostgreSQL
kubectl port-forward -n benchmark svc/benchmark-postgresql 5432:5432
```
