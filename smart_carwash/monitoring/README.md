# Smart Carwash Monitoring Setup

## 📊 Обзор

Система мониторинга Smart Carwash включает:

- **Grafana Loki** - централизованный сбор и анализ логов (14 дней хранения)
- **Prometheus** - сбор метрик системы и приложения
- **Grafana** - визуализация логов и метрик
- **Promtail** - агент сбора логов
- **Node Exporter** - системные метрики
- **PostgreSQL Exporter** - метрики базы данных

## 🚀 Запуск мониторинга

### 1. Запуск всех сервисов
```bash
cd /home/artem/smart_carwash/smart_carwash
docker-compose up -d
```

### 2. Проверка статуса сервисов
```bash
docker-compose ps
```

### 3. Просмотр логов
```bash
# Логи всех сервисов
docker-compose logs -f

# Логи конкретного сервиса
docker-compose logs -f grafana
docker-compose logs -f prometheus
docker-compose logs -f loki
```

## 🌐 Доступ к интерфейсам

### Grafana (Визуализация)
- **URL**: http://localhost:3000
- **Логин**: admin
- **Пароль**: admin

### Prometheus (Метрики)
- **URL**: http://localhost:9090

### Loki (Логи)
- **URL**: http://localhost:3100

## 📈 Дашборды Grafana

### 1. Smart Carwash - Logs Dashboard
- Обзор всех логов по сервисам
- Фильтрация по уровню (ERROR, WARN, INFO)
- Поиск по ключевым словам
- Встроенная инструкция по работе с логами

### 2. Smart Carwash - System Metrics
- CPU, Memory, Disk usage
- Network traffic
- Load average
- PostgreSQL connections

### 3. Smart Carwash - Application Metrics
- HTTP requests/sec
- Response times
- Error rates
- Session metrics
- Payment metrics
- Queue size

## 🔍 Работа с логами

### Основные запросы Loki

#### Поиск по сервису
```
{job="backend"}
{job="frontend"}
{job="nginx"}
{job="postgres"}
```

#### Поиск по уровню логирования
```
{job="backend", level="error"}
{job="backend", level="warn"}
{job="backend", level="info"}
```

#### Поиск по статусу HTTP
```
{job="nginx", status="200"}
{job="nginx", status=~"4..|5.."}
```

#### Поиск по мета-параметрам
```
{job="backend"} | json | user_id="12345"
{job="backend"} | json | session_id="abc-def"
{job="backend"} | json | payment_id="pay_123"
```

### Полезные метрики
```
# Счетчики логов
sum(rate({job="backend"}[5m]))
sum(count_over_time({job="backend"}[1h]))

# Ошибки по сервисам
sum(rate({job="backend", level="error"}[5m])) by (service)
```

## 📊 Метрики Prometheus

### Системные метрики
- `node_cpu_seconds_total` - использование CPU
- `node_memory_MemTotal_bytes` - общая память
- `node_filesystem_size_bytes` - размер диска
- `node_load1` - средняя нагрузка

### Метрики приложения
- `http_requests_total` - общее количество HTTP запросов
- `http_request_duration_seconds` - время выполнения запросов
- `active_connections` - активные соединения
- `sessions_total` - общее количество сессий
- `payments_total` - общее количество платежей
- `queue_size` - размер очереди
- `errors_total` - общее количество ошибок

### Метрики базы данных
- `pg_stat_database_numbackends` - активные соединения с БД
- `pg_stat_database_tup_returned` - возвращенные строки
- `pg_stat_database_tup_fetched` - полученные строки

## 🚨 Алерты

Система включает следующие алерты:

### Системные алерты
- **HighCPUUsage** - CPU > 80% более 5 минут
- **HighMemoryUsage** - память > 85% более 5 минут
- **HighDiskUsage** - диск > 90% более 5 минут
- **HighLoadAverage** - нагрузка > 4 более 5 минут

### Алерты приложения
- **HighHTTPErrorRate** - ошибки HTTP > 5% более 5 минут
- **HighHTTPResponseTime** - время ответа > 2 сек более 5 минут
- **TooManyActiveConnections** - соединения > 100 более 5 минут

### Алерты базы данных
- **HighDatabaseConnections** - соединения БД > 80 более 5 минут
- **DatabaseDown** - БД недоступна более 1 минуты

### Алерты сервисов
- **BackendDown** - backend недоступен более 1 минуты
- **FrontendDown** - frontend недоступен более 1 минуты
- **NginxDown** - nginx недоступен более 1 минуты

### Алерты мониторинга
- **PrometheusDown** - Prometheus недоступен более 1 минуты
- **LokiDown** - Loki недоступен более 1 минуты
- **GrafanaDown** - Grafana недоступен более 1 минуты

### Бизнес-алерты
- **HighQueueSize** - размер очереди > 10 более 5 минут
- **HighErrorRate** - ошибки > 0.1/сек более 5 минут
- **NoSessionsCreated** - нет сессий за последний час
- **NoPaymentsProcessed** - нет платежей за последний час

## 🔧 Конфигурация

### Структура файлов
```
monitoring/
├── loki-config.yml              # Конфигурация Loki
├── promtail-config.yml          # Конфигурация Promtail
├── prometheus.yml               # Конфигурация Prometheus
├── prometheus-rules/
│   └── alerts.yml               # Правила алертов
├── grafana-datasources/
│   └── datasources.yml          # Источники данных Grafana
└── grafana-dashboards/
    ├── dashboards.yml           # Конфигурация дашбордов
    ├── logs-dashboard.json      # Дашборд логов
    ├── system-metrics-dashboard.json  # Дашборд системных метрик
    └── app-metrics-dashboard.json     # Дашборд метрик приложения
```

### Порты
- **Grafana**: 3000
- **Prometheus**: 9090
- **Loki**: 3100
- **Node Exporter**: 9100
- **PostgreSQL Exporter**: 9187

### Volumes
- `loki_data` - данные Loki (14 дней)
- `prometheus_data` - данные Prometheus (30 дней)
- `grafana_data` - данные Grafana

## 🛠️ Troubleshooting

### Проблемы с запуском
1. Проверьте доступность портов
2. Убедитесь, что Docker и Docker Compose установлены
3. Проверьте права доступа к файлам конфигурации

### Проблемы с логами
1. Проверьте, что Promtail собирает логи
2. Убедитесь, что Loki получает данные
3. Проверьте конфигурацию Promtail

### Проблемы с метриками
1. Проверьте, что Prometheus собирает метрики
2. Убедитесь, что exporters работают
3. Проверьте конфигурацию Prometheus

### Проблемы с дашбордами
1. Проверьте подключение к источникам данных
2. Убедитесь, что дашборды загружены
3. Проверьте права доступа в Grafana

## 📝 Логирование в приложении

### Структурированные логи (JSON)
```go
logger.Info("User login", map[string]interface{}{
    "user_id": "12345",
    "ip": "192.168.1.1",
    "user_agent": "Mozilla/5.0...",
})
```

### Мета-параметры
- `service` - название сервиса
- `method` - HTTP метод
- `path` - путь запроса
- `status` - HTTP статус
- `duration` - время выполнения
- `user_id` - ID пользователя
- `session_id` - ID сессии
- `payment_id` - ID платежа
- `error` - описание ошибки

## 🔄 Обновление

### Обновление конфигурации
1. Остановите сервисы: `docker-compose down`
2. Обновите конфигурационные файлы
3. Запустите сервисы: `docker-compose up -d`

### Обновление дашбордов
1. Обновите JSON файлы дашбордов
2. Перезапустите Grafana: `docker-compose restart grafana`

## 📚 Дополнительные ресурсы

- [Grafana Loki Documentation](https://grafana.com/docs/loki/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Promtail Documentation](https://grafana.com/docs/loki/latest/clients/promtail/)

