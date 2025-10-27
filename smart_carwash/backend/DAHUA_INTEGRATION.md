# Интеграция с камерой Dahua (ITSAPI XML)

## Описание

Интеграция с камерой Dahua через протокол ITSAPI для автоматического завершения сессий при выезде автомобилей. Камера отправляет webhook в XML формате при распознавании номерных знаков, система автоматически завершает активные сессии БЕЗ частичного возврата.

**Поддерживаемые форматы:**
- ✅ **ITSAPI XML** (основной формат)
- ✅ **JSON** (для обратной совместимости)

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

### Webhook от камеры Dahua (ITSAPI XML)
```
POST /api/v1/dahua/anpr-webhook
Content-Type: application/xml
```
**⚠️ ВНИМАНИЕ: Аутентификация отключена для тестирования**

**Входящий XML от камеры Dahua (ITSAPI формат):**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<EventNotificationAlert>
    <ipAddress>192.168.1.100</ipAddress>
    <portNo>80</portNo>
    <protocol>HTTP</protocol>
    <macAddress>00:11:22:33:44:55</macAddress>
    <channelID>1</channelID>
    <dateTime>2025-10-20T14:05:33</dateTime>
    <activePostCount>1</activePostCount>
    <eventType>ANPR</eventType>
    <eventState>active</eventState>
    <eventDescription>ANPR Event</eventDescription>
    <licensePlate>A123BC77</licensePlate>
    <confidence>98</confidence>
    <direction>out</direction>
    <imagePath>http://192.168.1.100/pic/20251020_140533.jpg</imagePath>
</EventNotificationAlert>
```

**XML Ответ (ITSAPI формат):**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <result>OK</result>
    <message>Сессия успешно завершена для автомобиля A123BC77</message>
</Response>
```

### Webhook от камеры Dahua (JSON - обратная совместимость)
```
POST /api/v1/dahua/anpr-webhook
Content-Type: application/json
```
**⚠️ ВНИМАНИЕ: Аутентификация отключена для тестирования**

**Входящий JSON от камеры Dahua (реальная структура):**
```json
{
  "Picture": {
    "Plate": {
      "BoundingBox": [931, 1189, 1194, 1245],
      "Channel": 0,
      "IsExist": true,
      "PlateColor": "White",
      "PlateNumber": "O874yA154",
      "PlateType": "",
      "Region": "RUS",
      "UploadNum": 2
    },
    "SnapInfo": {
      "AllowUser": false,
      "AllowUserEndTime": "",
      "DefenceCode": "yCIWKTnP8IqWvxW2",
      "DeviceID": "d48a125d-abaf-9e67-2a55-58f94629abaf",
      "InCarPeopleNum": 0,
      "LanNo": 1,
      "OpenStrobe": false
    },
    "Vehicle": {
      "VehicleBoundingBox": [703, 552, 1459, 1315],
      "VehicleSeries": "Другой"
    }
  }
}
```

**JSON Ответ (для обратной совместимости):**
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

### Device Registration (ITSAPI)
```
POST /api/v1/NotificationInfo/DeviceInfo
Content-Type: application/json
```

**Входящий JSON от камеры (регистрация устройства):**
```json
{
  "DeviceID": "d48a125d-abaf-9e67-2a55-58f94629abaf",
  "DeviceModel": "DHI-ITC413-PW4D-IZ1(868MHz)",
  "DeviceName": "AM0C5C6PAJ23A56",
  "DeviceType": "Tollgate",
  "IPAddress": "192.168.88.100",
  "IPv6Address": "",
  "MACAddress": "30:dd:aa:62:a2:ae",
  "Manufacturer": "Dahua"
}
```

**JSON Ответ на регистрацию:**
```json
{
  "Result": "OK",
  "DeviceID": "d48a125d-abaf-9e67-2a55-58f94629abaf",
  "Message": "Device registered successfully",
  "Timestamp": "2025-10-20T14:05:33+08:00",
  "ServerID": "carwash-server-001",
  "Status": "Online"
}
```

### KeepAlive Heartbeat (ITSAPI)
```
GET/POST /api/v1/NotificationInfo/KeepAlive
Content-Type: application/xml
```

**XML Ответ на KeepAlive:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<HeartbeatResponse>
    <result>OK</result>
    <timestamp>2025-10-20T14:05:33+08:00</timestamp>
    <status>online</status>
    <nextHeartbeatInterval>30</nextHeartbeatInterval>
</HeartbeatResponse>
```

**JSON Ответы (для обратной совместимости):**
```json
{
  "success": true,
  "message": "Device info received",
  "timestamp": "2025-10-20T14:05:33+08:00",
  "status": "online"
}
```

```json
{
  "success": true,
  "message": "Heartbeat received",
  "timestamp": "2025-10-20T14:05:33+08:00",
  "status": "online",
  "nextHeartbeatInterval": 30
}
```

## Процесс регистрации устройства

### Автоматическая регистрация камеры
1. **Камера отправляет данные** - JSON с информацией об устройстве на `/NotificationInfo/DeviceInfo`
2. **Система парсит данные** - извлекает DeviceID, модель, IP, MAC и другие параметры
3. **Логирование регистрации** - детальная информация о камере сохраняется в логах
4. **Подтверждение регистрации** - система отправляет JSON ответ с подтверждением
5. **Статус "Online"** - камера получает статус "Online" и может отправлять события

### Поля регистрации устройства
- `DeviceID` - уникальный идентификатор устройства
- `DeviceModel` - модель камеры (например, DHI-ITC413-PW4D-IZ1)
- `DeviceName` - имя устройства в системе
- `DeviceType` - тип устройства (Tollgate, Camera, etc.)
- `IPAddress` - IP адрес камеры
- `MACAddress` - MAC адрес камеры
- `Manufacturer` - производитель (Dahua)

### Ответ системы на регистрацию
- `Result` - результат регистрации (OK/Error)
- `DeviceID` - подтверждение ID устройства
- `Message` - сообщение о статусе регистрации
- `Timestamp` - время подтверждения
- `ServerID` - идентификатор сервера
- `Status` - статус устройства (Online/Offline)

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

### Настройка камеры Dahua (ITSAPI XML)

1. **Вход в веб-интерфейс камеры**
2. **Event Settings → HTTP Events**
3. **Настройка HTTP Events для ITSAPI XML:**
   - URL: `http://your-server.com/api/v1/dahua/anpr-webhook`
   - Method: `POST`
   - Authentication: `Basic Auth`
   - Username: `dahua_user`
   - Password: `dahua_password`
   - Content-Type: `application/xml`
   - Format: `XML` (выберите XML формат)

4. **Настройка ANPR Events:**
   - Включить ANPR события
   - Настроить распознавание номерных знаков
   - Указать направление движения (in/out)
   - Убедиться, что камера отправляет события в формате ITSAPI XML

### Настройка камеры Dahua (JSON - для обратной совместимости)

1. **Вход в веб-интерфейс камеры**
2. **Event Settings → HTTP Events**
3. **Настройка HTTP Events для JSON:**
   - URL: `http://your-server.com/api/v1/dahua/anpr-webhook`
   - Method: `POST`
   - Authentication: `Basic Auth`
   - Username: `dahua_user`
   - Password: `dahua_password`
   - Content-Type: `application/json`
   - Format: `JSON` (выберите JSON формат)

4. **Настройка ANPR Events:**
   - Включить ANPR события
   - Настроить распознавание номерных знаков
   - Указать направление движения (in/out)

## Технические детали XML формата ITSAPI

### Структура XML запроса
Камера Dahua отправляет XML в формате ITSAPI со следующими полями:

- `ipAddress` - IP адрес камеры
- `portNo` - Порт камеры
- `protocol` - Протокол (обычно HTTP)
- `macAddress` - MAC адрес камеры
- `channelID` - ID канала камеры
- `dateTime` - Время события в формате ISO 8601
- `activePostCount` - Количество активных постов
- `eventType` - Тип события (ANPR)
- `eventState` - Состояние события (active/inactive)
- `eventDescription` - Описание события
- `licensePlate` - Распознанный номер автомобиля
- `confidence` - Уровень уверенности распознавания (0-100)
- `direction` - Направление движения (in/out)
- `imagePath` - Путь к изображению (опционально)

### Структура XML ответа
Система возвращает XML ответ в формате ITSAPI:

- `result` - Результат обработки (OK/ERROR)
- `message` - Сообщение о результате

### Автоматическое определение формата
Система автоматически определяет формат входящих данных по заголовку `Content-Type`:
- `application/xml` или `text/xml` → XML парсинг (ITSAPI)
- `application/json` → JSON парсинг (обратная совместимость)

### Структурированное логирование
Все запросы от камер Dahua логируются с детальной информацией для диагностики:

**Для DeviceInfo запросов (регистрация устройства):**
```
📨 Запрос на /NotificationInfo/DeviceInfo
📋 Method: POST
📋 Headers: map[Content-Type:[application/json] ...]
📄 Body: {"DeviceID":"d48a125d-abaf-9e67-2a55-58f94629abaf",...}
📋 Query params: map[]
📋 Content-Type: application/json
📋 Client IP: 192.168.88.100
📡 === РЕГИСТРАЦИЯ КАМЕРЫ DAHUA ===
🆔 Device ID: d48a125d-abaf-9e67-2a55-58f94629abaf
📷 Модель: DHI-ITC413-PW4D-IZ1(868MHz)
📛 Имя устройства: AM0C5C6PAJ23A56
🏷️ Тип: Tollgate
🌐 IP адрес: 192.168.88.100
🔧 MAC адрес: 30:dd:aa:62:a2:ae
🏭 Производитель: Dahua
✅ Камера успешно зарегистрирована: AM0C5C6PAJ23A56 (192.168.88.100)
```

**Для KeepAlive запросов:**
```
💓 Heartbeat от камеры на /NotificationInfo/KeepAlive
📋 Method: GET
📋 Client IP: 192.168.1.100
💓 Heartbeat body: <xml>...</xml> (если POST)
```

**Для ANPR Webhook:**
```
🚗 ANPR Webhook от камеры Dahua
📋 Method: POST
📋 Headers: map[Content-Type:[application/xml] ...]
📄 Body: <?xml version="1.0" encoding="UTF-8"?><EventNotificationAlert><licensePlate>A123BC77</licensePlate><direction>out</direction><confidence>98</confidence></EventNotificationAlert>
📋 Query params: map[]
📋 Content-Type: application/xml
📋 Client IP: 192.168.1.100
📋 License Plate: A123BC77
📋 Direction: out
📋 Confidence: 98
✅ ANPR событие обработано успешно: Сессия успешно завершена для автомобиля A123BC77
👤 Пользователь найден: A123BC77
🎯 Активная сессия найдена: 12345678-1234-1234-1234-123456789abc
```

## Безопасность

### Аутентификация
**⚠️ ВНИМАНИЕ: Аутентификация для ANPR webhook настроена только на IP whitelist**

1. **ANPR Webhook** - только IP whitelist (без Basic Auth)
2. **DeviceInfo/KeepAlive** - БЕЗ аутентификации (стандарт ITSAPI)
3. **Health Check** - БЕЗ аутентификации (для мониторинга)

### IP Whitelist
- **Включен для ANPR webhook** - проверка IP адреса камеры
- **Настройка через переменную окружения** `DAHUA_ALLOWED_IPS`
- **Поддержка CIDR** - можно указывать подсети
- **Пример**: `DAHUA_ALLOWED_IPS=192.168.88.100,192.168.1.0/24`

### Поддержка форматов
- ✅ **ITSAPI XML**: основной формат для камер Dahua
- ✅ **JSON**: обратная совместимость с существующими интеграциями
- ✅ **Автоматическое определение**: по Content-Type заголовку

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

