# Структурированное логирование

## Обзор

Приложение использует структурированное JSON-логирование с автоматическим trace_id для отслеживания запросов.

## Архитектура

- **Logger**: Пользовательский логгер на основе logrus с hook для записи в файл и stdout
- **Файл логов**: `/var/log/backend/app.log`
- **Promtail**: Собирает логи из файла и отправляет в Loki
- **Loki**: Хранит и индексирует логи
- **Grafana**: Визуализация и поиск логов

## Использование в коде

### В Handlers (с gin.Context)

```go
func (h *Handler) someHandler(c *gin.Context) {
    // Автоматически добавит trace_id из middleware
    logger.WithContext(c).Info("Processing request")
    logger.WithContext(c).Errorf("Error occurred: %v", err)
}
```

### В Service/Repository (без context)

```go
func (s *service) SomeMethod() {
    // Обычное логирование без trace_id
    logger.Info("Operation started")
    logger.WithFields(logrus.Fields{
        "user_id": userID,
        "amount": amount,
    }).Info("Payment processed")
}
```

### С явным trace_id

```go
func backgroundJob(traceID string) {
    logger.WithTraceID(traceID).Info("Background job started")
}
```

## Поля логов

Каждый лог содержит:
- `level` - уровень логирования (info, error, warn, debug)
- `message` - сообщение лога
- `time` - временная метка
- `trace_id` - уникальный ID запроса (автоматически для HTTP запросов)
- `func` - имя функции
- `file` - файл и строка

Дополнительные поля для HTTP запросов:
- `method` - HTTP метод
- `path` - путь запроса
- `status_code` - код ответа
- `duration` - время выполнения в мс
- `ip` - IP клиента
- `user_agent` - User-Agent
- `handler` - имя handler функции

## Поиск логов в Grafana

### Открыть Grafana Explore

1. Перейти в Grafana: http://localhost:3000
2. Выбрать "Explore" в левом меню
3. Выбрать источник данных "Loki"

### Примеры запросов

#### Все логи backend
```logql
{service="backend"}
```

#### Логи по trace_id (отследить весь запрос)
```logql
{service="backend", trace_id="123e4567-e89b-12d3-a456-426614174000"}
```

#### Только ошибки
```logql
{service="backend", level="error"}
```

#### Ошибки конкретного метода
```logql
{service="backend", method="POST"} |= "error"
```

#### Логи конкретного handler
```logql
{service="backend", handler=~".*createPayment.*"}
```

#### Логи с конкретным status code
```logql
{service="backend", status_code="500"}
```

#### Комбинированный поиск
```logql
{service="backend", level="error", method="POST"} |= "payment"
```

#### Поиск по тексту в сообщении
```logql
{service="backend"} |= "Payment failed"
```

#### Статистика ошибок
```logql
sum(count_over_time({service="backend", level="error"}[5m])) by (handler)
```

## Middleware

Logging middleware автоматически:
- Генерирует уникальный trace_id для каждого запроса
- Логирует начало и завершение запроса
- Добавляет trace_id в gin.Context
- Извлекает информацию о запросе (method, path, status, duration, etc.)

## Метрики

Помимо логов, доступны метрики Prometheus:
- HTTP запросы по статусам
- Время ответа
- Активные соединения
- Бизнес-метрики (сессии, платежи, очередь)

Метрики доступны на: http://localhost:9090

## Troubleshooting

### Логи не появляются в Grafana

1. Проверить, что backend пишет логи:
```bash
docker exec carwash_backend cat /var/log/backend/app.log
```

2. Проверить статус Promtail:
```bash
docker logs carwash_promtail
```

3. Проверить, что Loki получает логи:
```bash
curl http://localhost:3100/ready
```

### Нет trace_id в логах

Убедитесь, что в коде используется `logger.WithContext(c)` в handlers.

### Логи не структурированы

Проверьте, что логгер инициализирован с JSON форматом в `logger.Init()`.

