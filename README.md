# Smart Carwash Telegram Mini App

Telegram Mini App для умных автомоек с возможностью просмотра статуса боксов и управления ими.

## Технологии

### Backend
- Python 3.12
- FastAPI
- PostgreSQL
- SQLAlchemy
- Alembic
- python-telegram-bot

### Frontend
- React 18
- Telegram Mini App SDK (@tma.js/sdk)
- Styled Components
- Axios

### Инфраструктура
- Docker & Docker Compose
- Nginx
- Let's Encrypt для SSL

## Структура проекта

```
smart_carwash/
├── .env                    # Файл с переменными окружения
├── docker-compose.yml      # Конфигурация Docker Compose
├── backend/                # Бэкенд на Python/FastAPI
│   ├── app/                # Код приложения
│   │   ├── models/         # Модели данных
│   │   ├── routes/         # API маршруты
│   │   ├── services/       # Бизнес-логика
│   │   └── utils/          # Утилиты
│   ├── alembic/            # Миграции базы данных
│   ├── Dockerfile          # Dockerfile для бэкенда
│   └── requirements.txt    # Зависимости Python
├── frontend/               # Фронтенд на React
│   ├── public/             # Публичные файлы
│   ├── src/                # Исходный код
│   │   ├── components/     # React компоненты
│   │   ├── pages/          # Страницы приложения
│   │   ├── services/       # Сервисы для работы с API
│   │   └── utils/          # Утилиты
│   ├── Dockerfile          # Dockerfile для фронтенда
│   └── package.json        # Зависимости NPM
└── nginx/                  # Конфигурация Nginx
    ├── default.conf        # Конфигурация Nginx
    ├── Dockerfile          # Dockerfile для Nginx
    └── init-letsencrypt.sh # Скрипт для настройки SSL
```

## Установка и запуск

### Предварительные требования

- Docker и Docker Compose
- Доступ к серверу с публичным IP
- Зарегистрированный Telegram бот

### Настройка переменных окружения

1. Скопируйте файл `.env.example` в `.env` и заполните необходимые переменные:

```bash
cp .env.example .env
```

2. Отредактируйте файл `.env`:

```
# PostgreSQL Configuration
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres_password
POSTGRES_DB=carwash_db
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

# Backend Configuration
BACKEND_HOST=0.0.0.0
BACKEND_PORT=8000
DATABASE_URL=postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
TELEGRAM_BOT_NAME=your_telegram_bot_name
SERVER_IP=your_server_ip

# Frontend Configuration
REACT_APP_API_URL=https://${SERVER_IP}/api
REACT_APP_TELEGRAM_BOT_NAME=${TELEGRAM_BOT_NAME}

# Nginx Configuration
DOMAIN=${SERVER_IP}
```

### Запуск приложения

1. Соберите и запустите контейнеры:

```bash
docker-compose up -d
```

2. Проверьте, что все контейнеры запущены:

```bash
docker-compose ps
```

3. Проверьте логи:

```bash
docker-compose logs -f
```

### Настройка Telegram Mini App

1. Откройте BotFather в Telegram: https://t.me/BotFather
2. Отправьте команду `/mybots` и выберите вашего бота
3. Выберите "Bot Settings" > "Menu Button" > "Configure menu button"
4. Установите URL вашего приложения: `https://your_server_ip`
5. Отправьте команду `/setdomain` и укажите домен вашего сервера

## Разработка

### Локальный запуск бэкенда

```bash
cd backend
pip install -r requirements.txt
uvicorn app.main:app --reload
```

### Локальный запуск фронтенда

```bash
cd frontend
npm install
npm start
```

## API Endpoints

### Carwash API

- `GET /api/carwash/info` - Получить информацию о мойке
- `GET /api/carwash/boxes/{box_id}` - Получить информацию о конкретном боксе
- `PATCH /api/carwash/boxes/{box_id}` - Обновить статус бокса

### Users API

- `POST /api/users` - Создать нового пользователя
- `GET /api/users/{telegram_id}` - Получить информацию о пользователе
- `POST /api/users/ensure` - Создать пользователя, если он не существует

## Telegram Bot Commands

- `/start` - Начать работу с ботом и открыть Mini App
- `/help` - Показать справку

## Лицензия

MIT
