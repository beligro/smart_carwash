# Modbus HTTP Server

Отдельный HTTP сервер для интеграции с мойкой по протоколу Modbus TCP.

## Описание

Этот сервер принимает HTTP запросы и перенаправляет их на устройство Modbus TCP по локальной сети. Он служит прокси между HTTP API и Modbus TCP протоколом.

## API Endpoints

### Основные операции

- `POST /api/v1/modbus/coil` - запись значения в coil
- `POST /api/v1/modbus/light` - управление светом в боксе
- `POST /api/v1/modbus/chemistry` - управление химией в боксе
- `POST /api/v1/modbus/test-connection` - тестирование соединения
- `POST /api/v1/modbus/test-coil` - тестирование записи в coil

### Health Check

- `GET /health` - проверка состояния сервера

## Конфигурация

Создайте файл `config.env`:

```env
# HTTP сервер
SERVER_PORT=8081

# Modbus TCP
MODBUS_ENABLED=true
MODBUS_HOST=192.168.1.100
MODBUS_PORT=502

# Логирование
LOG_LEVEL=info
```

## Запуск

### Локальная разработка

```bash
# Установка зависимостей
go mod tidy

# Запуск сервера
go run cmd/main.go
```

### Docker

```bash
# Сборка образа
docker build -t modbus-server .

# Запуск контейнера
docker run -p 8081:8081 --env-file config.env modbus-server
```

### Docker Compose

```bash
# Запуск с docker-compose
docker-compose up -d
```

## Примеры запросов

### Управление светом

```bash
curl -X POST http://localhost:8081/api/v1/modbus/light \
  -H "Content-Type: application/json" \
  -d '{
    "box_id": "123e4567-e89b-12d3-a456-426614174000",
    "value": true
  }'
```

### Управление химией

```bash
curl -X POST http://localhost:8081/api/v1/modbus/chemistry \
  -H "Content-Type: application/json" \
  -d '{
    "box_id": "123e4567-e89b-12d3-a456-426614174000",
    "value": false
  }'
```

### Тест соединения

```bash
curl -X POST http://localhost:8081/api/v1/modbus/test-connection \
  -H "Content-Type: application/json" \
  -d '{
    "box_id": "123e4567-e89b-12d3-a456-426614174000"
  }'
```

## Ответы сервера

Все операции возвращают JSON ответ в формате:

```json
{
  "success": true,
  "message": "Операция выполнена успешно"
}
```

При ошибках:

```json
{
  "success": false,
  "message": "Описание ошибки"
}
```

## Особенности

- **Retry механизм**: При ошибках соединения выполняются 3 попытки с интервалом 2 секунды
- **Логирование**: Подробное логирование всех операций с box_id для отслеживания
- **Конфигурация**: Возможность отключения Modbus через MODBUS_ENABLED=false
- **Таймауты**: Настроенные таймауты для Modbus TCP соединений

## Развертывание

Сервер предназначен для развертывания на отдельной виртуальной машине рядом с ПЛК мойки для минимизации задержек сети.

### Переменные окружения для продакшена

```env
SERVER_PORT=8081
MODBUS_ENABLED=true
MODBUS_HOST=192.168.1.100  # IP адрес ПЛК
MODBUS_PORT=502
LOG_LEVEL=info
```
