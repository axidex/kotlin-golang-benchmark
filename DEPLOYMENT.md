# Deployment Guide

## Быстрый старт для k3s

### 1. Установка chart

```bash
# Установить приложение
cd charts/benchmark
helm install benchmark . -n benchmark --create-namespace
```

### 2. Настройка доступа (NodePort для локального k3s)

Если у вас локальный k3s без внешнего домена, используйте NodePort:

```bash
helm install benchmark . -n benchmark --create-namespace \
  --set ingress.enabled=false \
  --set quarkus.service.type=NodePort \
  --set golang.service.type=NodePort
```

После установки узнайте порты:

```bash
# Quarkus NodePort
kubectl get svc -n benchmark benchmark-quarkus -o jsonpath='{.spec.ports[0].nodePort}'

# Golang NodePort
kubectl get svc -n benchmark benchmark-golang -o jsonpath='{.spec.ports[0].nodePort}'
```

Доступ к приложениям:
- Quarkus: `http://localhost:<QUARKUS_NODEPORT>/api/products`
- Golang: `http://localhost:<GOLANG_NODEPORT>/api/products`

### 3. Настройка доступа (с доменом)

Если у вас есть домен:

```bash
helm install benchmark . -n benchmark --create-namespace \
  --set ingress.hosts[0].host=quarkus.yourdomain.com \
  --set ingress.hosts[1].host=golang.yourdomain.com \
  --set ingress.tls=null  # убрать TLS если нет cert-manager
```

### 4. Проверка работы

```bash
# Проверить статус подов
kubectl get pods -n benchmark

# Проверить логи
kubectl logs -n benchmark -l app=quarkus -f
kubectl logs -n benchmark -l app=golang -f

# Проверить сервисы
kubectl get svc -n benchmark

# Port-forward для локального доступа
kubectl port-forward -n benchmark svc/benchmark-quarkus 8080:8080
kubectl port-forward -n benchmark svc/benchmark-golang 8081:8080
```

### 5. Тестирование API

```bash
# Создать продукт (Quarkus)
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Product","description":"Test","price":100.0,"quantity":10}'

# Получить все продукты
curl http://localhost:8080/api/products

# Создать продукт (Golang)
curl -X POST http://localhost:8081/api/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Product","description":"Test","price":100.0,"quantity":10}'

# Получить все продукты
curl http://localhost:8081/api/products
```

### 6. Удаление

```bash
helm uninstall benchmark -n benchmark
kubectl delete namespace benchmark
```

## Кастомизация values.yaml

Основные параметры для изменения:

```yaml
# Изменить образы
quarkus:
  image:
    repository: axidex/benchmark-kotlin-quarkus
    tag: v1.0.0  # вместо latest

golang:
  image:
    repository: axidex/benchmark-golang-gin
    tag: v1.0.0  # вместо latest

# Изменить ресурсы
quarkus:
  resources:
    limits:
      cpu: 2000m
      memory: 1Gi

# Изменить пароль БД
postgresql:
  auth:
    password: "your-secure-password"
```

## Мониторинг и метрики

Приложения экспортируют метрики для Prometheus:
- Quarkus: `http://localhost:30080/q/metrics`
- Golang: `http://localhost:30081/metrics`

### Интеграция с kube-prometheus-stack

Если у вас установлен kube-prometheus-stack в namespace monitoring:

```bash
# Установка kube-prometheus-stack (если еще не установлен)
helm install mon prometheus-community/kube-prometheus-stack -n monitoring --create-namespace
```

1. **ServiceMonitor'ы** автоматически создадутся при установке chart и настроят scraping метрик

2. **Проверка что метрики собираются**:
```bash
# Проверить ServiceMonitors
kubectl get servicemonitor -n benchmark

# Port-forward к Prometheus и проверить targets
kubectl port-forward -n monitoring svc/mon-kube-prometheus-stack-prometheus 9090:9090
# Открыть http://localhost:9090/targets - должны быть benchmark-quarkus и benchmark-golang
```

3. **Доступ к Grafana**:
```bash
# Port-forward к Grafana
export POD_NAME=$(kubectl --namespace monitoring get pod -l "app.kubernetes.io/name=grafana,app.kubernetes.io/instance=mon" -oname)
kubectl --namespace monitoring port-forward $POD_NAME 3000
```

4. **Dashboard**:
   - Dashboard автоматически доступен если Grafana настроена на автоимпорт (sidecar)
   - Или импортировать вручную из `grafana/provisioning/dashboards/benchmark-dashboard.json`

### Если ServiceMonitor не подхватывается

Проверьте что Prometheus мониторит namespace benchmark:

```bash
# Добавить label к namespace
kubectl label namespace benchmark monitoring=enabled

# Проверить конфигурацию Prometheus Operator
kubectl get prometheus -n monitoring -o yaml | grep -A 5 serviceMonitorNamespaceSelector
```

Если нужно, обновите kube-prometheus-stack чтобы мониторить все namespaces:
```bash
helm upgrade mon prometheus-community/kube-prometheus-stack -n monitoring \
  --set prometheus.prometheusSpec.serviceMonitorNamespaceSelector={} \
  --set prometheus.prometheusSpec.serviceMonitorSelector={}
```

## Troubleshooting

### Приложение не стартует

```bash
# Проверить логи
kubectl logs -n benchmark deployment/benchmark-quarkus
kubectl logs -n benchmark deployment/benchmark-golang

# Проверить события
kubectl get events -n benchmark --sort-by='.lastTimestamp'

# Проверить описание пода
kubectl describe pod -n benchmark -l app=quarkus
```

### PostgreSQL не доступен

```bash
# Проверить PostgreSQL
kubectl logs -n benchmark -l app.kubernetes.io/name=postgresql

# Проверить service
kubectl get svc -n benchmark benchmark-postgresql

# Проверить credentials
kubectl get secret -n benchmark benchmark-postgresql -o yaml
```

### Ingress не работает

```bash
# Проверить Traefik (встроенный в k3s)
kubectl get pods -n kube-system -l app.kubernetes.io/name=traefik

# Проверить ingress
kubectl describe ingress -n benchmark

# Проверить logs Traefik
kubectl logs -n kube-system -l app.kubernetes.io/name=traefik
```
