# Интеграция Tinkoff Kassa в Smart Carwash

## Обзор

Данная документация описывает интеграцию платежной системы Tinkoff Kassa в Telegram Mini App для управления очередью автомойки.

## Архитектура

### Доменная модель

Интеграция следует принципам DDD (Domain-Driven Design):

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Session       │    │   Payment       │    │   Settings      │
│   Domain        │    │   Domain        │    │   Domain        │
│                 │    │                 │    │                 │
│ - Create        │◄──►│ - Calculate     │◄──►│ - Get Pricing   │
│ - Update Status │    │ - Create        │    │ - Update        │
│ - Process Queue │    │ - Handle Webhook│    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Новый Flow с платежами

1. **Выбор услуги** → Расчет цены через Payment Service
2. **Создание сессии** → Создание платежа в Tinkoff + сессии в БД
3. **Оплата** → Переход на страницу Tinkoff
4. **Webhook** → Обновление статуса платежа и сессии
5. **Очередь** → Worker обрабатывает сессии со статусом "in_queue"

## База данных

### Новые таблицы

#### Таблица `payments`
```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL,
    amount INTEGER NOT NULL, -- сумма в копейках
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    payment_url TEXT,
    tinkoff_id VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);
```

#### Обновление таблицы `sessions`
```sql
ALTER TABLE sessions ADD COLUMN payment_id UUID REFERENCES payments(id);
```

### Новые статусы сессий

- `created` - Создана (ожидает оплаты)
- `in_queue` - Оплачено, в очереди
- `payment_failed` - Ошибка оплаты
- `assigned` - Назначена на бокс
- `active` - Активна
- `complete` - Завершена
- `canceled` - Отменена
- `expired` - Истекла

## API Endpoints

### Payment Domain

#### Расчет цены
```
POST /payments/calculate-price
{
  "service_type": "wash",
  "with_chemistry": true,
  "rental_time_minutes": 10
}
```

#### Создание платежа
```
POST /payments/create
{
  "session_id": "uuid",
  "amount": 5000,
  "currency": "RUB"
}
```

#### Получение статуса платежа
```
GET /payments/status?payment_id=uuid
```

#### Повторная попытка оплаты
```
POST /payments/retry
{
  "session_id": "uuid"
}
```

#### Webhook от Tinkoff
```
POST /payments/webhook
{
  "TerminalKey": "...",
  "OrderId": "...",
  "Success": true,
  "Status": "CONFIRMED",
  "PaymentId": 123456789,
  "Amount": 5000,
  "Signature": "..."
}
```

### Session Domain (обновленный)

#### Создание сессии с платежом
```
POST /sessions/with-payment
{
  "user_id": "uuid",
  "service_type": "wash",
  "with_chemistry": true,
  "car_number": "A123BC",
  "rental_time_minutes": 10,
  "idempotency_key": "unique-key"
}
```

## Конфигурация

### Переменные окружения

```bash
# Tinkoff Kassa
TINKOFF_TERMINAL_KEY=your_terminal_key
TINKOFF_SECRET_KEY=your_secret_key
TINKOFF_SUCCESS_URL=https://t.me/your_bot?startapp=payment_success
TINKOFF_FAIL_URL=https://t.me/your_bot?startapp=payment_fail
```

### Настройки цен

Цены настраиваются в таблице `service_settings`. Миграция `000012_add_pricing_settings.up.sql` автоматически добавляет настройки цен при запуске:

```sql
INSERT INTO service_settings (service_type, setting_key, setting_value) VALUES 
    ('wash', 'price_per_minute', '3000'),           -- 30 руб/мин
    ('wash', 'chemistry_price_per_minute', '2000'), -- 20 руб/мин за химию
    ('air_dry', 'price_per_minute', '2000'),        -- 20 руб/мин
    ('air_dry', 'chemistry_price_per_minute', '1000'), -- 10 руб/мин за химию
    ('vacuum', 'price_per_minute', '1500'),         -- 15 руб/мин
    ('vacuum', 'chemistry_price_per_minute', '500');  -- 5 руб/мин за химию
```

**Важно:** Если настройки цен отсутствуют в базе данных, расчет цены будет завершаться ошибкой с понятным сообщением.

## Frontend Integration

### Новые компоненты

#### PriceCalculator
- Рассчитывает и отображает цену услуги
- Показывает разбивку цены (базовая + химия)
- Автоматически обновляется при изменении параметров

#### PaymentPage
- Страница оплаты с информацией о сессии
- Кнопка перехода к Tinkoff
- Автоматическая проверка статуса платежа
- Возможность повторной оплаты

### Обновленный Flow

1. **ServiceSelector** → Выбор услуги + расчет цены
2. **Создание сессии** → `POST /sessions/with-payment`
3. **PaymentPage** → Отображение платежа + переход к оплате
4. **Tinkoff** → Оплата на внешней странице
5. **Webhook** → Обновление статуса
6. **SessionDetails** → Отображение активной сессии

## Безопасность

### Валидация webhook'ов
- Проверка подписи HMAC-SHA256
- Валидация всех обязательных полей
- Логирование всех webhook'ов

### Идемпотентность
- Уникальные ключи для каждой сессии
- Проверка существующих сессий перед созданием
- Атомарные транзакции

### Обработка ошибок
- Детальное логирование всех операций
- Graceful degradation при недоступности Tinkoff
- Retry механизм для неудачных платежей

## Мониторинг

### Логирование
- Создание/обновление платежей
- Webhook события
- Ошибки интеграции
- Статистика платежей

### Метрики
- Количество успешных/неудачных платежей
- Время обработки webhook'ов
- Конверсия платежей

## Развертывание

### Миграции
```bash
# Применение миграций
migrate -path migrations -database "postgres://..." up

# Откат миграций
migrate -path migrations -database "postgres://..." down
```

### Конфигурация Tinkoff
1. Создать терминал в личном кабинете Tinkoff
2. Настроить webhook URL: `https://your-domain.com/api/payments/webhook`
3. Добавить переменные окружения
4. Протестировать интеграцию в тестовом режиме

## Тестирование

### Unit тесты
- Расчет цены для разных услуг
- Создание и обновление платежей
- Валидация webhook'ов

### Integration тесты
- Полный flow от выбора услуги до оплаты
- Обработка webhook'ов
- Обработка ошибок

### E2E тесты
- Тестирование в Telegram Mini App
- Проверка UI компонентов
- Тестирование реальных платежей (тестовый режим)

## Будущие улучшения

### Планируемые функции
1. **Продление сессии** - оплата дополнительного времени
2. **Автоматический возврат** - при досрочном завершении
3. **Полный возврат** - при отмене сессии
4. **Промокоды** - система скидок
5. **Подписки** - ежемесячные планы

### Архитектурные улучшения
1. **Event sourcing** - для аудита платежей
2. **Saga pattern** - для сложных транзакций
3. **Circuit breaker** - для надежности интеграции
4. **Rate limiting** - защита от спама 