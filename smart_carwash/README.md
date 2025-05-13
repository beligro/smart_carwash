# Умная автомойка - Telegram Mini App

Проект "Умная автомойка" представляет собой Telegram бота с Mini App для управления автомойкой.

## Технологии

### Бэкенд
- Golang
- Gin Web Framework
- GORM (ORM для PostgreSQL)
- Telegram Bot API

### Фронтенд
- React
- Telegram Mini App SDK
- Styled Components
- Axios

### Инфраструктура
- Docker
- Nginx
- PostgreSQL

## Структура проекта

```
smart_carwash/
├── .env                      # Файл окружения
├── docker-compose.yml        # Конфигурация Docker Compose
├── Makefile                  # Команды для управления проектом
├── backend/                  # Бэкенд на Golang
│   ├── cmd/                  # Точка входа приложения
│   ├── internal/             # Внутренний код приложения
│   │   ├── config/           # Конфигурация
│   │   ├── models/           # Модели данных
│   │   ├── handlers/         # Обработчики HTTP запросов
│   │   ├── repository/       # Работа с базой данных
│   │   ├── service/          # Бизнес-логика
│   │   └── telegram/         # Логика Telegram бота
│   └── migrations/           # Миграции базы данных
├── frontend/                 # Фронтенд на React
│   ├── public/               # Публичные файлы
│   └── src/                  # Исходный код React
│       ├── components/       # React компоненты
│       ├── services/         # Сервисы для работы с API
│       └── App.js            # Главный компонент
└── nginx/                    # Конфигурация Nginx
```

## Запуск проекта

### Предварительные требования

- Docker и Docker Compose
- Make (опционально, для использования Makefile)
- OpenSSL (для генерации SSL сертификатов)

### Шаги для запуска

1. Клонировать репозиторий:
   ```bash
   git clone <url-репозитория>
   cd smart_carwash
   ```

2. Создать SSL сертификаты для разработки:
   ```bash
   make ssl-dev
   ```

3. Собрать и запустить проект:
   ```bash
   make build
   make run
   ```

4. Применить миграции базы данных:
   ```bash
   make migrate
   ```

5. Проверить работу приложения:
   - Бэкенд API: https://158.160.105.190/api/wash-info
   - Фронтенд: https://158.160.105.190
   - Telegram бот: https://t.me/carwash_grom_test_bot

### Управление проектом

- Запуск: `make run`
- Остановка: `make stop`
- Перезапуск: `make restart`
- Просмотр логов: `make logs`
- Очистка: `make clean`

## API эндпоинты

- `GET /api/wash-info` - Получение информации о мойке (список боксов с их статусами)
- `POST /api/users` - Создание пользователя (при выполнении команды /start в Telegram боте)

## Telegram бот

Бот поддерживает следующие команды:
- `/start` - Начать работу с ботом и открыть Mini App

## Разработка

### Бэкенд

Для разработки бэкенда:
```bash
cd backend
go run cmd/main.go
```

### Фронтенд

Для разработки фронтенда:
```bash
cd frontend
npm install
npm start
```
