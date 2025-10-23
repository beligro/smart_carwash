# Интеграция с камерой Dahua

## Описание

Интеграция с камерой Dahua для автоматического завершения сессий при выезде автомобилей. Камера отправляет webhook при распознавании номерных знаков, система автоматически завершает активные сессии БЕЗ частичного возврата.

## Архитектура

### DDD домен `dahua`
```
backend/internal/domain/dahua/
├── models/
│   └── dahua.go              # Модели для webhook
├── service/
│   └── service.go            # Бизнес-логика
├── handlers/
│   ├── handlers.go           # HTTP обработчики
│   └── routes.go             # Маршруты
└── middleware/
    └── auth.go               # Basic Auth + IP whitelist
```

### Новые методы в session service
- `CompleteSessionWithoutRefund(sessionID uuid.UUID) error` - завершение сессии БЕЗ возврата
- `GetActiveSessionByUserID(userID uuid.UUID) (*models.Session, error)` - поиск активной сессии

### Новые методы в user service
- `GetUserByCarNumber(carNumber string) (*models.User, error)` - поиск пользователя по номеру

## API Endpoints

### Webhook от камеры Dahua
```
POST /api/v1/dahua/anpr-webhook
Authorization: Basic <base64(username:password)>
```

**Входящий JSON от камеры:**
```json
{
  "licensePlate": "A123BC77",
  "confidence": 98,
  "direction": "out",
  "eventType": "ANPR",
  "captureTime": "2025-10-20T14:05:33",
  "imagePath": "http://<camera_ip>/pic/20251020_140533.jpg"
}
```

**Ответ:**
```json
{
  "success": true,
  "message": "Сессия успешно завершена для автомобиля A123BC77"
}
```

### Health Check
```
GET /api/v1/dahua/health
```

## Логика обработки

1. **Проверка аутентификации**: Basic Auth (username/password) + IP whitelist
2. **Валидация данных**: проверка JSON формата и направления
3. **Фильтрация событий**: обрабатываются только события с `direction = "out"`
4. **Поиск пользователя**: по номеру автомобиля в поле `car_number`
5. **Поиск активной сессии**: пользователя со статусом `active`
6. **Завершение сессии**: БЕЗ частичного возврата через `CompleteSessionWithoutRefund()`
7. **Выключение оборудования**: автоматическое через Modbus (свет + химия)

## Конфигурация

### Переменные окружения
```env
# Аутентификация webhook
DAHUA_WEBHOOK_USERNAME=dahua_user
DAHUA_WEBHOOK_PASSWORD=dahua_password

# IP whitelist (через запятую, поддерживает CIDR)
DAHUA_ALLOWED_IPS=192.168.1.100,192.168.1.101,10.0.0.0/24
```

### Настройка камеры Dahua

1. **Вход в веб-интерфейс камеры**
2. **Event Settings → HTTP Events**
3. **Настройка HTTP Events:**
   - URL: `http://your-server.com/api/v1/dahua/anpr-webhook`
   - Method: `POST`
   - Authentication: `Basic Auth`
   - Username: `dahua_user`
   - Password: `dahua_password`
   - Content-Type: `application/json`

4. **Настройка ANPR Events:**
   - Включить ANPR события
   - Настроить распознавание номерных знаков
   - Указать направление движения (in/out)

## Безопасность

### Двойная аутентификация
1. **Basic Authentication**: username/password из переменных окружения
2. **IP Whitelist**: список разрешенных IP адресов

### Поддерживаемые форматы IP
- Точные IP: `192.168.1.100`
- CIDR подсети: `10.0.0.0/24`
- Список через запятую: `192.168.1.100,192.168.1.101,10.0.0.0/24`

### Заголовки прокси
Система автоматически определяет реальный IP клиента через заголовки:
- `X-Forwarded-For`
- `X-Real-IP`
- `RemoteAddr`

## Логирование

Все операции логируются с мета-параметрами:
- `dahua_license_plate` - номер автомобиля
- `dahua_direction` - направление движения
- `dahua_confidence` - уровень уверенности
- `dahua_client_ip` - IP адрес камеры
- `dahua_username` - имя пользователя аутентификации

## Обработка ошибок

### Сценарии обработки
1. **Неверная аутентификация** → HTTP 401
2. **IP не в whitelist** → HTTP 403
3. **Неверный JSON** → HTTP 400
4. **Direction != "out"** → HTTP 200 (игнорируется)
5. **Пользователь не найден** → HTTP 200 (игнорируется)
6. **Активная сессия не найдена** → HTTP 200 (игнорируется)
7. **Ошибка завершения сессии** → HTTP 500

### Логирование ошибок
Все ошибки логируются с контекстом для отладки и мониторинга.

## Мониторинг

### Health Check
```
GET /api/v1/dahua/health
```

**Ответ:**
```json
{
  "success": true,
  "message": "Dahua интеграция работает",
  "service": "dahua"
}
```

### Метрики
- Завершение сессий через Dahua
- Ошибки аутентификации
- Ошибки обработки webhook'ов

## Тестирование

### Тестовый запрос
```bash
curl -X POST http://localhost:8080/api/v1/dahua/anpr-webhook \
  -H "Content-Type: application/json" \
  -H "Authorization: Basic ZGFodWFfdXNlcjpkYWh1YV9wYXNzd29yZA==" \
  -d '{
    "licensePlate": "A123BC77",
    "confidence": 98,
    "direction": "out",
    "eventType": "ANPR",
    "captureTime": "2025-10-20T14:05:33",
    "imagePath": "http://192.168.1.100/pic/20251020_140533.jpg"
  }'
```

### Проверка health check
```bash
curl http://localhost:8080/api/v1/dahua/health
```

## Развертывание

1. **Добавить переменные окружения** в `.env` файл
2. **Перезапустить backend** для загрузки конфигурации
3. **Настроить камеру Dahua** согласно инструкции
4. **Проверить health check** для подтверждения работы
5. **Протестировать webhook** с тестовым запросом

## Особенности

### Без частичного возврата
При завершении сессии через Dahua webhook **НЕ выполняется** частичный возврат средств, так как клиент фактически использовал услугу.

### Автоматическое выключение оборудования
- Выключение света в боксе через Modbus
- Выключение химии (если была включена) через Modbus
- Обновление статуса бокса

### Обработка только выездов
Система обрабатывает только события с `direction = "out"`. События въезда игнорируются.

## Troubleshooting

### Проблемы с аутентификацией
1. Проверить переменные окружения `DAHUA_WEBHOOK_USERNAME` и `DAHUA_WEBHOOK_PASSWORD`
2. Проверить настройки Basic Auth в камере
3. Проверить логи на ошибки аутентификации

### Проблемы с IP whitelist
1. Проверить переменную `DAHUA_ALLOWED_IPS`
2. Проверить реальный IP камеры в логах
3. Добавить IP камеры в whitelist

### Проблемы с обработкой
1. Проверить, что пользователь существует с указанным номером
2. Проверить, что у пользователя есть активная сессия
3. Проверить логи на ошибки завершения сессии

### Проблемы с Modbus
1. Проверить настройки Modbus в конфигурации
2. Проверить доступность Modbus устройства
3. Проверить логи на ошибки Modbus операций

