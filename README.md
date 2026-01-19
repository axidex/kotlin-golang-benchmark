# Kotlin Quarkus vs Golang Gin Benchmark

Проект для сравнения производительности CRUD приложений на Quarkus (Kotlin) и Gin (Golang) с использованием PostgreSQL, Prometheus и Grafana.

## Архитектура

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

## Быстрый старт

### 1. Запуск всех сервисов

```bash
docker-compose up --build
```

Это запустит:
- **PostgreSQL** (порт 5432) - общая БД для обоих приложений
- **Quarkus App** (порт 8080) - Kotlin CRUD API
- **Golang App** (порт 8081) - Go CRUD API
- **Prometheus** (порт 9090) - сбор метрик
- **Grafana** (порт 3000) - визуализация метрик

### 2. Доступ к сервисам

| Сервис | URL | Credentials |
|--------|-----|-------------|
| Quarkus API | http://localhost:8080/api/products | - |
| Golang API | http://localhost:8081/api/products | - |
| Grafana | http://localhost:3000 | admin / admin |
| Prometheus | http://localhost:9090 | - |
| Quarkus Metrics | http://localhost:8080/q/metrics | - |
| Golang Metrics | http://localhost:8081/metrics | - |

### 3. Запуск бенчмарка

```bash
./benchmark.sh
```

## API Endpoints

Оба приложения предоставляют идентичный REST API:

- `GET /api/products` - получить все продукты
- `GET /api/products/:id` - получить продукт по ID
- `POST /api/products` - создать новый продукт
- `PUT /api/products/:id` - обновить продукт
- `DELETE /api/products/:id` - удалить продукт

### Примеры запросов

```bash
# Создать продукт в Quarkus
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","description":"Gaming laptop","price":1500.00,"quantity":10}'

# Создать продукт в Golang
curl -X POST http://localhost:8081/api/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Mouse","description":"Gaming mouse","price":50.00,"quantity":100}'

# Получить все продукты
curl http://localhost:8080/api/products
curl http://localhost:8081/api/products
```

## Метрики в Grafana

После запуска откройте Grafana по адресу http://localhost:3000 (admin/admin).

Дашборд "Kotlin Quarkus vs Golang Gin Benchmark" содержит:

### Панели производительности
- **Requests per Second** - количество запросов в секунду для каждого приложения
- **Average Response Time** - среднее время отклика в миллисекундах
- **HTTP Status Codes** - распределение HTTP кодов ответа

### Панели использования ресурсов
- **Memory Usage** - потребление памяти (Heap для JVM, Alloc для Go)
- **Total Requests** - общее количество обработанных запросов
- **Active Goroutines** - количество активных горутин (только для Go)

## Структура проекта

```
.
├── kotlin-quarkus/              # Quarkus приложение
│   ├── src/main/kotlin/
│   │   └── dev/sourcecraft/dolgintsev/
│   │       ├── entity/Product.kt
│   │       └── resource/ProductResource.kt
│   ├── src/main/resources/
│   │   └── application.properties
│   ├── build.gradle.kts
│   └── Dockerfile
│
├── golang-gin/                  # Golang приложение
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── models/product.go
│   │   ├── handlers/product_handler.go
│   │   ├── database/database.go
│   │   └── middleware/metrics.go
│   ├── go.mod
│   └── Dockerfile
│
├── grafana/                     # Конфигурация Grafana
│   └── provisioning/
│       ├── datasources/prometheus.yml
│       └── dashboards/
│           ├── dashboard.yml
│           └── benchmark-dashboard.json
│
├── prometheus.yml               # Конфигурация Prometheus
├── docker-compose.yml           # Docker compose для всех сервисов
└── benchmark.sh                 # Скрипт для бенчмарка
```

## Технологии

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

## Остановка и очистка

```bash
# Остановить все сервисы
docker-compose down

# Остановить и удалить volumes (очистить данные БД)
docker-compose down -v
```

## Benchmarking Tips

1. **Прогрев**: Запустите несколько запросов перед началом бенчмарка для прогрева JVM
2. **Нагрузка**: Варьируйте параметры `-n` (requests) и `-c` (concurrency) в benchmark.sh
3. **Мониторинг**: Наблюдайте за метриками в Grafana в реальном времени во время бенчмарка
4. **Ресурсы**: Убедитесь, что Docker выделено достаточно ресурсов (CPU, RAM)

## Результаты

После запуска бенчмарка сравните:
- Throughput (запросов в секунду)
- Latency (время отклика)
- Использование памяти
- Стабильность под нагрузкой