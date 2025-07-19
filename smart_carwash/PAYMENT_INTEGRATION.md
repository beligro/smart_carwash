# Интеграция Tinkoff Kassa в Telegram Mini App

## Обзор

Данная интеграция добавляет полноценную платежную систему в Telegram Mini App для автомойки с использованием Tinkoff Kassa.

## Архитектура

### Бэкенд (Golang)

Создан новый домен `payment` согласно DDD архитектуре:

```
backend/internal/domain/payment/
├── models/
│   ├── payment.go          # Модели платежей
│   └── tinkoff.go          # Модели Tinkoff API
├── repository/
│   └── repository.go       # Работа с БД
├── service/
│   └── service.go          # Бизнес-логика
├── handlers/
│   └── handlers.go         # API ручки
└── tinkoff/
    └── client.go           # HTTP клиент для Tinkoff API
```

### Фронтенд (React)

Добавлены компоненты для работы с платежами:

```
frontend/src/
├── shared/services/
│   └── PaymentService.js   # Сервис для работы с платежами
└── apps/telegram/components/
    ├── PaymentModal/       # Модальное окно платежей
    └── SessionExtension/   # Продление сессии с платежом
```

## Функциональность

### 1. Предоплата за очередь
- Пользователь выбирает услугу
- Система создает платеж в Tinkoff
- После успешной оплаты пользователь добавляется в очередь

### 2. Продление сессии
- Пользователь может продлить активную сессию
- Оплата производится за дополнительное время
- Сессия автоматически продлевается после оплаты

### 3. Автоматические возвраты
- При досрочном завершении сессии
- При отмене бронирования
- Система повторных попыток для неудачных возвратов

### 4. Идемпотентность
- Все операции защищены от дублирования
- Уникальные ключи для каждого платежа
- События платежей для аудита

## Настройка

### 1. Переменные окружения

Добавьте в `.env` файл:

```env
# Tinkoff Kassa
TINKOFF_TERMINAL_KEY=your_terminal_key
TINKOFF_SECRET_KEY=your_secret_key
TINKOFF_BASE_URL=https://securepay.tinkoff.ru

# URL для платежей
PAYMENT_SUCCESS_URL=https://your-domain.com/telegram?payment=success
PAYMENT_FAIL_URL=https://your-domain.com/telegram?payment=fail
PAYMENT_WEBHOOK_URL=https://your-domain.com/api/v1/payments/webhook
```

### 2. Миграции базы данных

Запустите миграцию для создания таблиц платежей:

```bash
make migrate-up
```

### 3. Настройка Tinkoff Kassa

1. Создайте аккаунт в Tinkoff Kassa
2. Получите Terminal Key и Secret Key
3. Настройте webhook URL в личном кабинете
4. Укажите URL для успешных и неудачных платежей

## API Endpoints

### Платежи

- `POST /api/v1/payments` - Создание платежа
- `GET /api/v1/payments/by-id?id={id}` - Получение платежа
- `GET /api/v1/payments/status?id={id}` - Статус платежа
- `POST /api/v1/payments/webhook` - Webhook от Tinkoff

### Возвраты

- `POST /api/v1/payments/refunds` - Создание возврата
- `POST /api/v1/payments/refunds/automatic?session_id={id}` - Автоматический возврат
- `POST /api/v1/payments/refunds/full?payment_id={id}` - Полный возврат

### Интеграция с другими доменами

- `POST /api/v1/payments/queue?user_id={id}&service_type={type}` - Платеж за очередь
- `POST /api/v1/payments/session-extension?session_id={id}&extension_minutes={minutes}` - Платеж за продление
- `GET /api/v1/payments/session?session_id={id}` - Проверка платежа сессии

### Административные

- `GET /admin/v1/payments?limit={limit}&offset={offset}` - Список платежей
- `GET /admin/v1/payments/by-id?id={id}` - Детали платежа
- `POST /admin/v1/payments/reconcile?date={date}` - Сверка платежей

## Безопасность

### 1. Идемпотентность
- Все платежи имеют уникальные ключи
- Проверка дублирования на уровне БД
- События для аудита операций

### 2. Валидация webhook'ов
- Проверка подписи от Tinkoff
- Валидация данных платежа
- Защита от повторной обработки

### 3. Обработка ошибок
- Логирование всех операций
- Повторные попытки для возвратов
- Graceful degradation при сбоях

## Мониторинг

### 1. Логирование
- Все платежные операции логируются
- Мета-данные для поиска по ID
- Ошибки с контекстом

### 2. Метрики
- Количество успешных/неудачных платежей
- Время обработки операций
- Статистика возвратов

### 3. Алерты
- Неудачные webhook'и
- Превышение лимита ошибок
- Проблемы с возвратами

## Тестирование

### 1. Тестовые данные
```bash
# Создание тестового платежа
curl -X POST "http://localhost:8080/api/v1/payments/queue" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test-user-id",
    "service_type": "wash"
  }'
```

### 2. Webhook тестирование
```bash
# Имитация webhook от Tinkoff
curl -X POST "http://localhost:8080/api/v1/payments/webhook" \
  -H "Content-Type: application/json" \
  -d '{
    "TerminalKey": "test",
    "OrderId": "test-order",
    "Success": true,
    "Status": "CONFIRMED",
    "PaymentId": "test-payment",
    "Amount": 5000
  }'
```

## Развертывание

### 1. Docker
```bash
# Сборка и запуск
docker-compose up --build
```

### 2. Переменные окружения
Убедитесь, что все переменные окружения настроены:
- `TINKOFF_TERMINAL_KEY`
- `TINKOFF_SECRET_KEY`
- `PAYMENT_WEBHOOK_URL`

### 3. SSL сертификат
Для продакшена необходим SSL сертификат для webhook'ов.

## Troubleshooting

### 1. Платеж не создается
- Проверьте Terminal Key и Secret Key
- Убедитесь в правильности URL'ов
- Проверьте логи на ошибки

### 2. Webhook не приходит
- Проверьте настройки webhook URL в Tinkoff
- Убедитесь в доступности сервера
- Проверьте SSL сертификат

### 3. Возвраты не работают
- Проверьте статус исходного платежа
- Убедитесь в правильности суммы
- Проверьте логи возвратов

## Дальнейшее развитие

### 1. Дополнительные функции
- Промокоды и скидки
- Подписки на услуги
- Лояльность клиентов

### 2. Интеграции
- Другие платежные системы
- CRM системы
- Аналитика платежей

### 3. Оптимизация
- Кэширование статусов платежей
- Асинхронная обработка
- Масштабирование 