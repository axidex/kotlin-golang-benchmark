# Deployment Guide

## Quick start for k3s

### 1. Install chart

```bash
# Install application
cd charts/benchmark
helm install benchmark . -n benchmark --create-namespace
```

### 2. Setup access (NodePort for local k3s)

If you have local k3s without external domain, use NodePort:

```bash
helm install benchmark . -n benchmark --create-namespace \
  --set ingress.enabled=false \
  --set quarkus.service.type=NodePort \
  --set golang.service.type=NodePort
```

After installation, find out the ports:

```bash
# Quarkus NodePort
kubectl get svc -n benchmark benchmark-quarkus -o jsonpath='{.spec.ports[0].nodePort}'

# Golang NodePort
kubectl get svc -n benchmark benchmark-golang -o jsonpath='{.spec.ports[0].nodePort}'
```

Access to applications:
- Quarkus: `http://localhost:<QUARKUS_NODEPORT>/api/products`
- Golang: `http://localhost:<GOLANG_NODEPORT>/api/products`

### 3. Setup access (with domain)

If you have a domain:

```bash
helm install benchmark . -n benchmark --create-namespace \
  --set ingress.hosts[0].host=quarkus.yourdomain.com \
  --set ingress.hosts[1].host=golang.yourdomain.com \
  --set ingress.tls=null  # remove TLS if no cert-manager
```

### 4. Check operation

```bash
# Check pod status
kubectl get pods -n benchmark

# Check logs
kubectl logs -n benchmark -l app=quarkus -f
kubectl logs -n benchmark -l app=golang -f

# Check services
kubectl get svc -n benchmark

# Port-forward for local access
kubectl port-forward -n benchmark svc/benchmark-quarkus 8080:8080
kubectl port-forward -n benchmark svc/benchmark-golang 8081:8080
```

### 5. API testing

```bash
# Create product (Quarkus)
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Product","description":"Test","price":100.0,"quantity":10}'

# Get all products
curl http://localhost:8080/api/products

# Create product (Golang)
curl -X POST http://localhost:8081/api/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Product","description":"Test","price":100.0,"quantity":10}'

# Get all products
curl http://localhost:8081/api/products
```

### 6. Removal

```bash
helm uninstall benchmark -n benchmark
kubectl delete namespace benchmark
```

## Customizing values.yaml

Main parameters to change:

```yaml
# Change images
quarkus:
  image:
    repository: axidex/benchmark-kotlin-quarkus
    tag: v1.0.0  # instead of latest

golang:
  image:
    repository: axidex/benchmark-golang-gin
    tag: v1.0.0  # instead of latest

# Change resources
quarkus:
  resources:
    limits:
      cpu: 2000m
      memory: 1Gi

# Change DB password
postgresql:
  auth:
    password: "your-secure-password"
```

## Monitoring and metrics

Applications export metrics for Prometheus:
- Quarkus: `http://localhost:30080/q/metrics`
- Golang: `http://localhost:30081/metrics`

### Integration with kube-prometheus-stack

If you have kube-prometheus-stack installed in monitoring namespace:

```bash
# Install kube-prometheus-stack (if not already installed)
helm install mon prometheus-community/kube-prometheus-stack -n monitoring --create-namespace
```

1. **ServiceMonitors** will be created automatically during chart installation and configure metrics scraping

2. **Check that metrics are being collected**:
```bash
# Check ServiceMonitors
kubectl get servicemonitor -n benchmark

# Port-forward to Prometheus and check targets
kubectl port-forward -n monitoring svc/mon-kube-prometheus-stack-prometheus 9090:9090
# Open http://localhost:9090/targets - should see benchmark-quarkus and benchmark-golang
```

3. **Access Grafana**:
```bash
# Port-forward to Grafana
export POD_NAME=$(kubectl --namespace monitoring get pod -l "app.kubernetes.io/name=grafana,app.kubernetes.io/instance=mon" -oname)
kubectl --namespace monitoring port-forward $POD_NAME 3000
```

4. **Dashboard**:
   - Dashboard automatically available if Grafana configured for auto-import (sidecar)
   - Or import manually from `grafana/provisioning/dashboards/benchmark-dashboard.json`

### If ServiceMonitor is not picked up

Check that Prometheus monitors benchmark namespace:

```bash
# Add label to namespace
kubectl label namespace benchmark monitoring=enabled

# Check Prometheus Operator configuration
kubectl get prometheus -n monitoring -o yaml | grep -A 5 serviceMonitorNamespaceSelector
```

If needed, update kube-prometheus-stack to monitor all namespaces:
```bash
helm upgrade mon prometheus-community/kube-prometheus-stack -n monitoring \
  --set prometheus.prometheusSpec.serviceMonitorNamespaceSelector={} \
  --set prometheus.prometheusSpec.serviceMonitorSelector={}
```

## Troubleshooting

### Application won't start

```bash
# Check logs
kubectl logs -n benchmark deployment/benchmark-quarkus
kubectl logs -n benchmark deployment/benchmark-golang

# Check events
kubectl get events -n benchmark --sort-by='.lastTimestamp'

# Check pod description
kubectl describe pod -n benchmark -l app=quarkus
```

### PostgreSQL not available

```bash
# Check PostgreSQL
kubectl logs -n benchmark -l app.kubernetes.io/name=postgresql

# Check service
kubectl get svc -n benchmark benchmark-postgresql

# Check credentials
kubectl get secret -n benchmark benchmark-postgresql -o yaml
```

### Ingress not working

```bash
# Check Traefik (built-in to k3s)
kubectl get pods -n kube-system -l app.kubernetes.io/name=traefik

# Check ingress
kubectl describe ingress -n benchmark

# Check Traefik logs
kubectl logs -n kube-system -l app.kubernetes.io/name=traefik
```
