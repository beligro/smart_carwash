# Telegram Mini App для умной автомойки

Telegram Mini App для умных автомоек, позволяющее пользователям проверять доступность моечных боксов.

## Функциональность

- Интеграция с Telegram ботом
- Telegram Mini App с современным интерфейсом (использует официальный SDK @telegram-apps/sdk-react)
- Информация о доступных боксах автомойки в реальном времени
- Управление пользователями с разделением на администраторов и обычных пользователей
- Настройка Docker и Nginx для простого развертывания

## Технологический стек

### Бэкенд
- Python с FastAPI
- База данных PostgreSQL
- ORM SQLAlchemy
- Библиотека Python Telegram Bot

### Фронтенд
- React с TypeScript
- Официальный SDK для Telegram Mini App (@telegram-apps/sdk-react)
- Styled Components
- Интеграция с Telegram Web App API

### Инфраструктура
- Docker и Docker Compose
- Nginx для обслуживания фронтенда и проксирования API-запросов

## Структура проекта

```
smart_carwash/
├── backend/                # Python бэкенд
│   ├── app/
│   │   ├── models/         # Модели базы данных и схемы
│   │   ├── routes/         # API маршруты
│   │   ├── utils/          # Вспомогательные функции
│   │   ├── bot.py          # Реализация Telegram бота
│   │   └── main.py         # FastAPI приложение
│   ├── Dockerfile
│   ├── requirements.txt
│   └── run.py              # Скрипт точки входа
├── frontend/               # React фронтенд
│   ├── public/
│   ├── src/
│   │   ├── components/     # React компоненты
│   │   ├── pages/          # Компоненты страниц
│   │   ├── services/       # API сервисы
│   │   └── utils/          # Вспомогательные функции
│   ├── Dockerfile
│   └── package.json
├── docker-compose.yml
└── .env.example
```

## Инструкции по установке

### Предварительные требования

- Docker и Docker Compose
- Токен Telegram бота (от BotFather)
- Домен для размещения Mini App (или localhost для разработки)

### Настройка

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/yourusername/smart_carwash.git
   cd smart_carwash
   ```

2. Создайте файлы окружения:
   ```bash
   cp .env.example .env
   cp backend/.env.example backend/.env
   cp frontend/.env.example frontend/.env
   ```

3. Отредактируйте файл `.env`, добавив токен вашего Telegram бота и другие настройки:
   ```
   TELEGRAM_BOT_TOKEN=ваш_токен_telegram_бота
   TELEGRAM_BOT_USERNAME=имя_вашего_бота
   WEBAPP_URL=https://ваш-домен.com
   ```

### Запуск приложения

1. Соберите и запустите Docker контейнеры:
   ```bash
   docker-compose up -d
   ```

2. Сервисы будут доступны по следующим адресам:
   - Фронтенд: http://localhost
   - API бэкенда: http://localhost:8000
   - PostgreSQL: localhost:5432

3. Для остановки приложения:
   ```bash
   docker-compose down
   ```

### Разработка

Для разработки вы можете запустить сервисы отдельно:

#### Бэкенд

```bash
cd backend
pip install -r requirements.txt
python run.py --service all
```

#### Фронтенд

```bash
cd frontend
npm install
npm start
```

## Создание Telegram Mini App

1. Напишите [@BotFather](https://t.me/botfather) в Telegram
2. Создайте нового бота или выберите существующего
3. Используйте команду `/newapp` для создания Mini App
4. Установите URL веб-приложения на URL вашего развернутого фронтенда
5. Настройте бота для отправки ссылки на Mini App пользователям

## Лицензия

MIT
