#!/bin/bash

# Smart Carwash Monitoring Stop Script

echo "🛑 Stopping Smart Carwash Monitoring System..."

# Проверяем, что мы в правильной директории
if [ ! -f "docker-compose.yml" ]; then
    echo "❌ Error: docker-compose.yml not found. Please run this script from the project root directory."
    exit 1
fi

# Останавливаем сервисы мониторинга
echo "🐳 Stopping monitoring services..."
docker-compose down

# Проверяем статус
echo "📊 Checking service status..."
docker-compose ps

echo ""
echo "✅ Monitoring system stopped successfully!"
echo ""
echo "💡 To start monitoring again: ./start-monitoring.sh"
echo "🗑️  To remove all data: docker-compose down -v"

