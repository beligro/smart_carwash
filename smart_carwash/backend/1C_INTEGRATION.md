# Интеграция с 1C для платежей через кассира

## Обзор

Система поддерживает интеграцию с внешней системой учета 1C для обработки платежей, совершенных через кассира. Платежи от 1C обрабатываются через webhook и автоматически создают сессии мойки с соответствующими платежами.

## Конфигурация

### Переменные окружения

Добавьте следующие переменные в ваш `.env` файл:

```env
# API ключ для аутентификации 1C webhook
API_KEY_1C=your-secret-api-key-here

# UUID пользователя-кассира в системе
CASHIER_USER_ID=12345678-1234-1234-1234-123456789abc
```

### Миграция базы данных

Примените новую миграцию для добавления поля `payment_method`:

```bash
migrate -path migrations -database "postgres://user:password@localhost:5432/carwash?sslmode=disable" up
```

## API Endpoint

### POST /api/v1/1c/payment-callback

Обрабатывает платежи от системы 1C.

#### Аутентификация

Требуется заголовок `Authorization` с форматом:
```
Authorization: Bearer your-secret-api-key-here
```

#### Request Body

```json
{
  "service_type": "wash",
  "with_chemistry": true,
  "payment_time": "2023-10-27T10:00:00Z",
  "amount": 50000,
  "rental_time_minutes": 20
}
```

#### Поля запроса

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| `service_type` | string | Да | Тип услуги: `wash`, `air_dry`, `vacuum` |
| `with_chemistry` | boolean | Нет | Использование химии (только для `wash`) |
| `payment_time` | string (ISO 8601) | Да | Время платежа |
| `amount` | integer | Да | Сумма в копейках |
| `rental_time_minutes` | integer | Да | Время мойки в минутах |

#### Response

**Успешный ответ (200 OK):**
```json
{
  "success": true,
  "session_id": "12345678-1234-1234-1234-123456789abc",
  "payment_id": "87654321-4321-4321-4321-cba987654321",
  "message": "Payment processed successfully"
}
```

**Ошибка валидации (400 Bad Request):**
```json
{
  "error": "Key: 'CashierPaymentRequest.ServiceType' Error:Field validation for 'ServiceType' failed on the 'oneof' tag"
}
```

**Ошибка аутентификации (401 Unauthorized):**
```json
{
  "error": "Неверный API ключ"
}
```

**Ошибка сервера (500 Internal Server Error):**
```json
{
  "error": "CASHIER_USER_ID не настроен"
}
```

## Бизнес-логика

### Создание сессии

1. Проверяется валидность входных данных
2. Создается сессия с параметрами:
   - `UserID`: значение из `CASHIER_USER_ID`
   - `Status`: `"in_queue"`
   - `CarNumber`: пустая строка
   - `ServiceType`: из запроса
   - `WithChemistry`: из запроса
   - `RentalTimeMinutes`: из запроса

### Создание платежа

1. Создается платеж с параметрами:
   - `SessionID`: ID созданной сессии
   - `Amount`: сумма из запроса
   - `Status`: `"succeeded"`
   - `PaymentType`: `"main"`
   - `PaymentMethod`: `"cashier"`
   - `TinkoffID`: пустая строка

### Валидация

- `service_type` должен быть одним из: `wash`, `air_dry`, `vacuum`
- `with_chemistry` разрешено только для `service_type: wash`
- `amount` должен быть положительным числом
- `rental_time_minutes` должен быть положительным числом

## Логирование

Все операции логируются с мета-параметрами:

```
Received 1C payment callback: ServiceType=wash, WithChemistry=true, Amount=50000, RentalTimeMinutes=20
Successfully processed 1C payment: SessionID=12345678-1234-1234-1234-123456789abc, PaymentID=87654321-4321-4321-4321-cba987654321, Amount=50000
```

## Безопасность

- API ключ передается через переменную окружения
- Все запросы аутентифицируются через middleware
- Валидация всех входных данных
- Логирование всех операций

## Примеры использования

### cURL запрос

```bash
curl -X POST https://h2o-nsk.ru/api/1c/payment-callback \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-secret-api-key-here" \
  -d '{
    "service_type": "wash",
    "with_chemistry": true,
    "payment_time": "2025-08-03T12:00:00Z",
    "amount": 50000,
    "rental_time_minutes": 20
  }'
```

### JavaScript

```javascript
const response = await fetch('/api/v1/1c/payment-callback', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer your-secret-api-key-here'
  },
  body: JSON.stringify({
    service_type: 'wash',
    with_chemistry: true,
    payment_time: '2023-10-27T10:00:00Z',
    amount: 50000,
    rental_time_minutes: 20
  })
});

const result = await response.json();
console.log(result);
```

## Обработка ошибок

### Частые ошибки

1. **CASHIER_USER_ID не настроен** - проверьте переменную окружения
2. **Неверный формат CASHIER_USER_ID** - убедитесь, что это валидный UUID
3. **Chemistry is not available for service type** - химия доступна только для мойки
4. **Invalid service type** - используйте только разрешенные типы услуг

### Мониторинг

- Проверяйте логи на наличие ошибок
- Мониторьте статус webhook'ов от 1C
- Отслеживайте созданные сессии и платежи в админке 