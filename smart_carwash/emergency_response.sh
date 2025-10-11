#!/bin/bash

echo "🚨 ЭКСТРЕННОЕ РЕАГИРОВАНИЕ НА КРИПТОМАЙНЕР"

# Останавливаем все контейнеры
echo "🛑 Останавливаю все контейнеры..."
docker-compose down

# Убиваем подозрительные процессы
echo "💀 Убиваю подозрительные процессы..."
pkill -f "kinsing"
pkill -f "kdevtmpfsi"
pkill -f "libsystem"
pkill -f "xmrig"
pkill -f "cpuminer"
pkill -f "ccminer"
pkill -f "minerd"

# Очищаем /tmp
echo "🧹 Очищаю /tmp..."
find /tmp -name "*kinsing*" -delete 2>/dev/null
find /tmp -name "*kdevtmpfsi*" -delete 2>/dev/null
find /tmp -name "*libsystem*" -delete 2>/dev/null
find /tmp -name "*xmrig*" -delete 2>/dev/null

# Очищаем Docker volumes
echo "🗑️ Удаляю зараженные Docker volumes..."
docker volume prune -f

# Пересоздаем контейнеры
echo "🔄 Пересоздаю контейнеры..."
docker-compose up -d postgres
sleep 10

# Восстанавливаем базу данных
echo "💾 Восстанавливаю базу данных..."
if [ -f "backups/backup_20250827_164501.sql" ]; then
    docker exec -i carwash_postgres psql -U postgres -d carwash < backups/backup_20250827_164501.sql
    echo "✅ База данных восстановлена"
else
    echo "⚠️ Резервная копия не найдена"
fi

# Запускаем все сервисы
echo "🚀 Запускаю все сервисы..."
docker-compose up -d

echo "✅ Экстренное реагирование завершено!"
echo "📊 Статус контейнеров:"
docker-compose ps
