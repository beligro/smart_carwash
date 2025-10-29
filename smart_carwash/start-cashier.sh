#!/bin/bash

# Скрипт для запуска фронтенда кассира
# Автор: Smart Carwash Team

set -e

echo "🚀 Запуск фронтенда кассира..."

# Проверяем, что мы в правильной директории
if [ ! -f "docker-compose.cashier.yml" ]; then
    echo "❌ Ошибка: docker-compose.cashier.yml не найден. Запустите скрипт из корневой директории проекта."
    exit 1
fi

# Проверяем наличие Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Ошибка: Docker не установлен или не доступен."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Ошибка: Docker Compose не установлен или не доступен."
    exit 1
fi

# Останавливаем существующий контейнер если он запущен
echo "🛑 Останавливаем существующий контейнер кассира..."
docker-compose -f docker-compose.cashier.yml down 2>/dev/null || true

# Собираем и запускаем контейнер
echo "🔨 Собираем образ кассира..."
docker-compose -f docker-compose.cashier.yml build --no-cache

echo "🚀 Запускаем контейнер кассира..."
docker-compose -f docker-compose.cashier.yml up -d

# Ждем запуска контейнера
echo "⏳ Ждем запуска контейнера..."
sleep 10

# Проверяем статус контейнера
if docker-compose -f docker-compose.cashier.yml ps | grep -q "Up"; then
    echo "✅ Фронтенд кассира успешно запущен!"
    echo "🌐 Доступен по адресу: http://localhost:3000"
    echo "🔐 Страница входа: http://localhost:3000/cashier/login"
    echo ""
    echo "🌍 Для доступа с других серверов используйте:"
    echo "   http://<IP_СЕРВЕРА>:3000"
    echo "   http://<IP_СЕРВЕРА>:3000/cashier/login"
    echo ""
    echo "📋 Полезные команды:"
    echo "  Просмотр логов: docker-compose -f docker-compose.cashier.yml logs -f"
    echo "  Остановка: docker-compose -f docker-compose.cashier.yml down"
    echo "  Перезапуск: docker-compose -f docker-compose.cashier.yml restart"
    echo ""
    echo "🔍 Проверка статуса:"
    docker-compose -f docker-compose.cashier.yml ps
else
    echo "❌ Ошибка при запуске контейнера кассира"
    echo "📋 Логи контейнера:"
    docker-compose -f docker-compose.cashier.yml logs
    exit 1
fi
