# Benchmark Helm Chart

Helm chart для развертывания Kotlin Quarkus и Golang Gin приложений в Kubernetes (k3s).

## Установка

### 1. Установить chart

```bash
# Базовая установка
helm install benchmark . -n benchmark --create-namespace

# С кастомными значениями
helm install benchmark . -n benchmark --create-namespace \
  --set ingress.hosts[0].host=quarkus.yourdomain.com \
  --set ingress.hosts[1].host=golang.yourdomain.com
```

### 4. Обновить существующий релиз

```bash
helm upgrade benchmark . -n benchmark
```

### 5. Удалить релиз

```bash
helm uninstall benchmark -n benchmark
```

## Конфигурация

### Основные параметры

| Параметр | Описание | Значение по умолчанию |
|----------|----------|----------------------|
| `quarkus.enabled` | Включить Quarkus приложение | `true` |
| `quarkus.image.repository` | Docker образ Quarkus | `axidex/benchmark-kotlin-quarkus` |
| `quarkus.image.tag` | Тег образа | `latest` |
| `golang.enabled` | Включить Golang приложение | `true` |
| `golang.image.repository` | Docker образ Golang | `axidex/benchmark-golang-gin` |
| `golang.image.tag` | Тег образа | `latest` |
| `postgresql.enabled` | Включить PostgreSQL | `true` |
| `postgresql.auth.database` | Имя БД | `benchmark` |
| `postgresql.auth.username` | Пользователь БД | `postgres` |
| `postgresql.auth.password` | Пароль БД | `postgres` |

### Ingress

По умолчанию используется Traefik (встроенный в k3s).

Измените хосты в `values.yaml`:

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

## Примеры использования

### Установка без TLS

```bash
helm install benchmark . -n benchmark --create-namespace \
  --set ingress.tls=null
```

### Использование LoadBalancer вместо Ingress

```bash
helm install benchmark . -n benchmark --create-namespace \
  --set ingress.enabled=false \
  --set quarkus.service.type=LoadBalancer \
  --set golang.service.type=LoadBalancer
```

### Увеличение ресурсов

```bash
helm install benchmark . -n benchmark --create-namespace \
  --set quarkus.resources.limits.memory=1Gi \
  --set quarkus.resources.limits.cpu=2000m
```

## Проверка статуса

```bash
# Проверить поды
kubectl get pods -n benchmark

# Проверить сервисы
kubectl get svc -n benchmark

# Проверить ingress
kubectl get ingress -n benchmark

# Логи Quarkus
kubectl logs -n benchmark -l app=quarkus -f

# Логи Golang
kubectl logs -n benchmark -l app=golang -f
```

## Endpoints

После установки приложения будут доступны по следующим адресам:

- Quarkus: `http://quarkus.example.com/api/products`
- Golang: `http://golang.example.com/api/products`
- Quarkus Health: `http://quarkus.example.com/q/health`
- Golang Health: `http://golang.example.com/health`
- Quarkus Metrics: `http://quarkus.example.com/q/metrics`
- Golang Metrics: `http://golang.example.com/metrics`

## Troubleshooting

### Проверить подключение к БД

```bash
kubectl exec -it -n benchmark deployment/benchmark-quarkus -- env | grep DATABASE
kubectl exec -it -n benchmark deployment/benchmark-golang -- env | grep DATABASE
```

### Проверить логи PostgreSQL

```bash
kubectl logs -n benchmark -l app.kubernetes.io/name=postgresql -f
```

### Port-forward для локального доступа

```bash
# Quarkus
kubectl port-forward -n benchmark svc/benchmark-quarkus 8080:8080

# Golang
kubectl port-forward -n benchmark svc/benchmark-golang 8081:8080

# PostgreSQL
kubectl port-forward -n benchmark svc/benchmark-postgresql 5432:5432
```