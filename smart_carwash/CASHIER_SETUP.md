# Настройка фронтенда кассира

Этот документ описывает, как запустить только фронтенд кассира локально через Docker.

## Требования

- Docker и Docker Compose установлены
- VPN подключение к серверу с бэкендом (10.0.0.4)
- Доступ к API по адресу: `http://10.0.0.4/api/v1`

## Быстрый запуск

```bash
# Запуск через скрипт (рекомендуется)
./start-cashier.sh

# Или вручную
docker-compose -f docker-compose.cashier.yml up --build -d
```

## Доступ к приложению

### Локальный доступ
- **Главная страница**: http://localhost:3000
- **Вход кассира**: http://localhost:3000/cashier/login
- **Интерфейс кассира**: http://localhost:3000/cashier

### Внешний доступ (с других серверов)
- **Главная страница**: http://<IP_СЕРВЕРА>:3000
- **Вход кассира**: http://<IP_СЕРВЕРА>:3000/cashier/login
- **Интерфейс кассира**: http://<IP_СЕРВЕРА>:3000/cashier

> Замените `<IP_СЕРВЕРА>` на реальный IP адрес сервера, где запущен контейнер

## Управление контейнером

```bash
# Просмотр логов
docker-compose -f docker-compose.cashier.yml logs -f

# Остановка
docker-compose -f docker-compose.cashier.yml down

# Перезапуск
docker-compose -f docker-compose.cashier.yml restart

# Пересборка
docker-compose -f docker-compose.cashier.yml up --build -d
```

## Структура файлов

- `docker-compose.cashier.yml` - конфигурация Docker Compose
- `frontend/Dockerfile.cashier` - Dockerfile для сборки кассира
- `frontend/nginx.cashier.conf` - конфигурация nginx
- `cashier.env` - переменные окружения
- `start-cashier.sh` - скрипт для запуска

## Настройка API

API URL настраивается в файле `docker-compose.cashier.yml`:

```yaml
environment:
  - REACT_APP_API_URL=http://10.0.0.4/api/v1
```

## Устранение неполадок

### Контейнер не запускается
```bash
# Проверьте логи
docker-compose -f docker-compose.cashier.yml logs

# Проверьте статус
docker-compose -f docker-compose.cashier.yml ps
```

### Проблемы с доступом
- Убедитесь, что порт 3000 открыт в файрволе
- Проверьте, что контейнер запущен: `docker ps | grep cashier`
- Проверьте логи: `docker-compose -f docker-compose.cashier.yml logs`

### Проблемы с API
- Убедитесь, что VPN подключен
- Проверьте доступность API: `curl http://10.0.0.4/api/v1/health`
- Проверьте настройки CORS на бэкенде

### Проблемы с портами
Если порт 3000 занят, измените в `docker-compose.cashier.yml`:
```yaml
ports:
  - "8080:80"  # Теперь доступен на http://localhost:8080
  # или
  - "9000:80"  # Теперь доступен на http://localhost:9000
```

## Особенности

- Фронтенд работает в standalone режиме только для кассира
- Использует nginx для раздачи статических файлов
- Настроен для работы через VPN к бэкенду
- Поддерживает SPA роутинг
- Оптимизирован для production использования
