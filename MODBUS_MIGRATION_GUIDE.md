# Руководство по миграции Modbus на отдельный сервер

## Обзор изменений

Modbus интеграция была выделена в отдельный HTTP сервер для улучшения архитектуры и возможности развертывания на отдельной машине.

## Новая архитектура

```
Основной бэкенд (smart_carwash)  →  HTTP  →  Modbus сервер  →  Modbus TCP  →  ПЛК
```

## Что изменилось

### 1. Новый Modbus сервер
- **Расположение**: `/modbus/` (рядом с smart_carwash)
- **Порт**: 8081
- **Функциональность**: Полностью дублирует текущую modbus логику
- **API**: HTTP endpoints для всех modbus операций

### 2. Основной бэкенд
- **HTTP клиент**: Взаимодействие с modbus сервером через HTTP
- **Интерфейс**: ModbusServiceInterface для совместимости
- **Адаптер**: ModbusAdapter для прозрачной замены

## Настройка

### 1. Конфигурация основного бэкенда

Добавьте в `.env` файл основного бэкенда:

```env
# Настройки Modbus HTTP сервера
MODBUS_SERVER_HOST=192.168.1.100  # IP адрес сервера с modbus
MODBUS_SERVER_PORT=8081
```

### 2. Запуск Modbus сервера

```bash
cd modbus/
docker-compose up -d
```

### 3. Проверка интеграции

```bash
# Проверка health check
curl http://192.168.1.100:8081/health

# Тест управления светом
curl -X POST http://192.168.1.100:8081/api/v1/modbus/light \
  -H "Content-Type: application/json" \
  -d '{"box_id": "test-uuid", "value": true}'
```

## API Endpoints Modbus сервера

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/v1/modbus/coil` | Запись в coil |
| POST | `/api/v1/modbus/light` | Управление светом |
| POST | `/api/v1/modbus/chemistry` | Управление химией |
| POST | `/api/v1/modbus/test-connection` | Тест соединения |
| POST | `/api/v1/modbus/test-coil` | Тест coil |
| GET | `/health` | Health check |

## Структура запросов

### Управление светом
```json
{
  "box_id": "uuid",
  "value": true
}
```

### Управление химией
```json
{
  "box_id": "uuid", 
  "value": false
}
```

### Запись в coil
```json
{
  "box_id": "uuid",
  "register": "0x001",
  "value": true
}
```

## Преимущества новой архитектуры

1. **Изоляция**: Modbus логика изолирована в отдельном сервисе
2. **Масштабируемость**: Возможность развертывания на отдельной машине
3. **Надежность**: HTTP клиент с retry механизмом
4. **Тестируемость**: Легкое тестирование без реального ПЛК
5. **Мониторинг**: Отдельное логирование modbus операций

## Обратная совместимость

- Session service продолжает работать без изменений
- Все методы интерфейса сохранены
- Логика продления времени сессии при ошибках сохранена

## Мониторинг

### Логи основного бэкенда
```bash
# Поиск modbus операций
grep "ModbusAdapter" logs/app.log
```

### Логи modbus сервера
```bash
# Мониторинг modbus операций
docker logs modbus-server -f
```

## Troubleshooting

### Проблема: Modbus сервер недоступен
```bash
# Проверка доступности
curl http://192.168.1.100:8081/health

# Проверка конфигурации
docker exec modbus-server cat /root/config.env
```

### Проблема: Ошибки HTTP клиента
- Проверьте настройки `MODBUS_SERVER_HOST` и `MODBUS_SERVER_PORT`
- Убедитесь, что modbus сервер запущен
- Проверьте сетевую связность

### Проблема: Modbus операции не выполняются
- Проверьте `MODBUS_ENABLED=true` в modbus сервере
- Проверьте настройки `MODBUS_HOST` и `MODBUS_PORT`
- Проверьте доступность ПЛК

## Развертывание в продакшене

### 1. Modbus сервер
```bash
# На отдельной машине рядом с ПЛК
docker-compose up -d
```

### 2. Основной бэкенд
```env
# В .env основного бэкенда
MODBUS_SERVER_HOST=192.168.1.100  # IP modbus сервера
MODBUS_SERVER_PORT=8081
```

### 3. Мониторинг
- Настройте мониторинг health check modbus сервера
- Настройте алерты на ошибки HTTP клиента
- Мониторинг логов modbus операций
