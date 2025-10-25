# Документация бэкенда

## Обзор

Бэкенд системы умной автомойки построен на Go с использованием архитектуры DDD (Domain-Driven Design). Система предоставляет API для управления автомойкой, включая боксы, сессии, очереди и пользователей.

## Архитектура

### DDD структура

Система разделена на домены, каждый из которых содержит 4 слоя:

1. **models** - модели доменной сущности
2. **repository** - слой для общения с базой данных
3. **service** - слой со всей продуктовой логикой
4. **handlers** - слой с API ручками

### Домены

- **auth** - авторизация и аутентификация
- **user** - управление пользователями
- **washbox** - управление боксами мойки
- **session** - управление сессиями мойки
- **queue** - управление очередью
- **settings** - настройки системы
- **telegram** - интеграция с Telegram

## API Endpoints

### Административные endpoints

#### Управление боксами мойки

**GET /admin/washboxes**
Получение списка боксов с фильтрацией

Query параметры:
- `status` - фильтр по статусу (free, busy, reserved, maintenance)
- `service_type` - фильтр по типу услуги (wash, air_dry, vacuum)
- `limit` - лимит записей (по умолчанию 50)
- `offset` - смещение (по умолчанию 0)

Response:
```json
{
  "wash_boxes": [
    {
      "id": "uuid",
      "number": 1,
      "status": "free",
      "service_type": "wash",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 10,
  "limit": 50,
  "offset": 0
}
```

**POST /admin/washboxes**
Создание нового бокса

Request body:
```json
{
  "number": 1,
  "status": "free",
  "service_type": "wash"
}
```

Response:
```json
{
  "wash_box": {
    "id": "uuid",
    "number": 1,
    "status": "free",
    "service_type": "wash",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**PUT /admin/washboxes**
Редактирование бокса

Request body:
```json
{
  "id": "uuid",
  "number": 2,
  "status": "maintenance",
  "service_type": "air_dry"
}
```

**DELETE /admin/washboxes**
Удаление бокса

Request body:
```json
{
  "id": "uuid"
}
```

**GET /admin/washboxes/by-id**
Получение бокса по ID

Query параметры:
- `id` - ID бокса (обязательный)

#### Управление сессиями мойки

**GET /admin/sessions**
Получение списка сессий с фильтрацией

Query параметры:
- `user_id` - фильтр по пользователю
- `box_id` - фильтр по боксу
- `status` - фильтр по статусу (created, assigned, active, complete, canceled, expired)
- `service_type` - фильтр по типу услуги (wash, air_dry, vacuum)
- `date_from` - фильтр по дате от (формат: YYYY-MM-DD)
- `date_to` - фильтр по дате до (формат: YYYY-MM-DD)
- `limit` - лимит записей (по умолчанию 50)
- `offset` - смещение (по умолчанию 0)

Response:
```json
{
  "sessions": [
    {
      "id": "uuid",
      "user_id": "uuid",
      "box_id": "uuid",
      "box_number": 1,
      "status": "active",
      "service_type": "wash",
      "with_chemistry": false,
      "rental_time_minutes": 5,
      "extension_time_minutes": 0,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "status_updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 100,
  "limit": 50,
  "offset": 0
}
```

**GET /admin/sessions/by-id**
Получение сессии по ID

Query параметры:
- `id` - ID сессии (обязательный)

#### Мониторинг очереди

**GET /admin/queue/status**
Получение статуса очереди

Query параметры:
- `include_details` - включить детальную информацию (true/false)

Response:
```json
{
  "queue_status": {
    "all_boxes": [...],
    "wash_queue": {
      "service_type": "wash",
      "boxes": [...],
      "queue_size": 5,
      "has_queue": true
    },
    "air_dry_queue": {...},
    "vacuum_queue": {...},
    "total_queue_size": 10,
    "has_any_queue": true
  },
  "details": {
    "users_in_queue": [
      {
        "user_id": "uuid",
        "username": "user123",
        "first_name": "Иван",
        "last_name": "Иванов",
        "service_type": "wash",
        "position": 1,
        "waiting_since": "2024-01-01T00:00:00Z"
      }
    ],
    "queue_order": ["uuid1", "uuid2", "uuid3"]
  }
}
```

#### Управление пользователями

**GET /admin/users**
Получение списка пользователей

Query параметры:
- `limit` - лимит записей (по умолчанию 50)
- `offset` - смещение (по умолчанию 0)

Response:
```json
{
  "users": [
    {
      "id": "uuid",
      "telegram_id": 123456789,
      "username": "user123",
      "first_name": "Иван",
      "last_name": "Иванов",
      "is_admin": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 50,
  "limit": 50,
  "offset": 0
}
```

**GET /admin/users/by-id**
Получение пользователя по ID

Query параметры:
- `id` - ID пользователя (обязательный)

### Пользовательские endpoints

#### Сессии

**POST /sessions**
Создание новой сессии

**GET /sessions**
Получение активной сессии пользователя

**GET /sessions/by-id**
Получение сессии по ID

**POST /sessions/start**
Запуск сессии

**POST /sessions/complete**
Завершение сессии

**POST /sessions/extend**
Продление сессии

**GET /sessions/history**
История сессий пользователя

#### Очередь

**GET /queue-status**
Получение статуса очереди и боксов

#### Пользователи

**POST /users**
Создание пользователя

**GET /users/by-telegram-id**
Получение пользователя по Telegram ID

## Модели данных

### WashBox

```go
type WashBox struct {
    ID          uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    Number      int            `json:"number" gorm:"uniqueIndex"`
    Status      string         `json:"status" gorm:"default:free"`
    ServiceType string         `json:"service_type" gorm:"default:wash"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
```

Статусы бокса:
- `free` - Свободен
- `reserved` - Зарезервирован
- `busy` - Занят
- `maintenance` - На обслуживании

Типы услуг:
- `wash` - Мойка
- `air_dry` - Обдув воздухом
- `vacuum` - Пылесос

### Session

```go
type Session struct {
    ID                           uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    UserID                       uuid.UUID      `json:"user_id" gorm:"index;type:uuid"`
    BoxID                        *uuid.UUID     `json:"box_id,omitempty" gorm:"index;type:uuid"`
    BoxNumber                    *int           `json:"box_number,omitempty" gorm:"-"`
    Status                       string         `json:"status" gorm:"default:created;index"`
    ServiceType                  string         `json:"service_type,omitempty" gorm:"default:null"`
    WithChemistry                bool           `json:"with_chemistry" gorm:"default:false"`
    RentalTimeMinutes            int            `json:"rental_time_minutes" gorm:"default:5"`
    ExtensionTimeMinutes         int            `json:"extension_time_minutes" gorm:"default:0"`
    IdempotencyKey               string         `json:"idempotency_key,omitempty" gorm:"index"`
    IsExpiringNotificationSent   bool           `json:"is_expiring_notification_sent" gorm:"default:false"`
    IsCompletingNotificationSent bool           `json:"is_completing_notification_sent" gorm:"default:false"`
    CreatedAt                    time.Time      `json:"created_at"`
    UpdatedAt                    time.Time      `json:"updated_at"`
    StatusUpdatedAt              time.Time      `json:"status_updated_at"`
    DeletedAt                    gorm.DeletedAt `json:"-" gorm:"index"`
}
```

Статусы сессии:
- `created` - Создана
- `assigned` - Назначена на бокс
- `active` - Активна
- `complete` - Завершена
- `canceled` - Отменена
- `expired` - Истек срок резервирования

### User

```go
type User struct {
    ID         uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    TelegramID int64          `json:"telegram_id" gorm:"uniqueIndex"`
    Username   string         `json:"username"`
    FirstName  string         `json:"first_name"`
    LastName   string         `json:"last_name"`
    IsAdmin    bool           `json:"is_admin" gorm:"default:false"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`
    DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}
```

## Бизнес-логика

### Обработка очереди

Система автоматически обрабатывает очередь каждые 5 секунд:

1. Проверяет сессии со статусом `created`
2. Находит свободные боксы соответствующего типа услуги
3. Назначает сессии на боксы
4. Обновляет статус сессии на `assigned`
5. Обновляет статус бокса на `reserved`

### Завершение сессий

Система автоматически завершает истекшие сессии:

1. Проверяет активные сессии
2. Вычисляет время окончания (время создания + время мойки + время продления)
3. Завершает сессии, время которых истекло
4. Освобождает боксы

### Уведомления

Система отправляет уведомления пользователям:

1. При истечении времени резервирования
2. При завершении сессии
3. При назначении на бокс

## Безопасность

### Авторизация

- JWT токены для аутентификации
- Проверка прав доступа для административных endpoints
- Валидация входных данных

### Валидация

Все входные данные валидируются с использованием тегов `binding`:

```go
type AdminCreateWashBoxRequest struct {
    Number      int    `json:"number" binding:"required"`
    Status      string `json:"status" binding:"required,oneof=free reserved busy maintenance"`
    ServiceType string `json:"service_type" binding:"required,oneof=wash air_dry vacuum"`
}
```

### Логирование

Все административные операции логируются с мета-параметрами:

```go
c.Set("meta", gin.H{
    "washbox_id": resp.WashBox.ID,
    "number":     resp.WashBox.Number,
})
```

## Конфигурация

### Переменные окружения

- `BACKEND_PORT` - порт сервера
- `DB_HOST` - хост базы данных
- `DB_PORT` - порт базы данных
- `DB_USER` - пользователь базы данных
- `DB_PASSWORD` - пароль базы данных
- `DB_NAME` - имя базы данных
- `TELEGRAM_BOT_TOKEN` - токен Telegram бота
- `JWT_SECRET` - секрет для JWT токенов

### База данных

Система использует PostgreSQL с миграциями:

```bash
# Применение миграций
migrate -path migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" up

# Откат миграций
migrate -path migrations -database "postgres://user:password@localhost:5432/dbname?sslmode=disable" down
```

## Развертывание

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### Запуск

```bash
# Установка зависимостей
go mod download

# Применение миграций
migrate -path migrations -database "$DATABASE_URL" up

# Запуск сервера
go run cmd/main.go
```

## Мониторинг

### Метрики

- Количество активных сессий
- Размер очереди
- Статус боксов
- Время обработки запросов

### Логи

- Логи ошибок
- Логи административных операций
- Логи производительности

## API версионирование

Текущая версия API: v1

Все endpoints находятся в корневом пути `/`. Версионирование планируется в будущих релизах.

## Ограничения

### Rate Limiting

- 100 запросов в минуту для пользовательских endpoints
- 1000 запросов в минуту для административных endpoints

### Размер данных

- Максимальный размер запроса: 10MB
- Максимальное количество записей в ответе: 1000

## Обработка ошибок

Все ошибки возвращаются в формате:

```json
{
  "error": "Описание ошибки"
}
```

HTTP статус коды:
- 200 - Успешный запрос
- 201 - Ресурс создан
- 400 - Неверный запрос
- 401 - Не авторизован
- 403 - Доступ запрещен
- 404 - Ресурс не найден
- 500 - Внутренняя ошибка сервера 