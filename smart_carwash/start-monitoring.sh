#!/bin/bash

# Smart Carwash Monitoring Startup Script

echo "🚀 Starting Smart Carwash Monitoring System..."

# Проверяем, что мы в правильной директории
if [ ! -f "docker-compose.yml" ]; then
    echo "❌ Error: docker-compose.yml not found. Please run this script from the project root directory."
    exit 1
fi

# Проверяем, что Docker запущен
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running. Please start Docker first."
    exit 1
fi

# Создаем необходимые директории для логов
echo "📁 Creating log directories..."
sudo mkdir -p /var/log/backend
sudo chmod 755 /var/log/backend

# Запускаем сервисы мониторинга
echo "🐳 Starting monitoring services..."
docker-compose up -d loki promtail prometheus node-exporter postgres-exporter grafana

# Ждем запуска сервисов
echo "⏳ Waiting for services to start..."
sleep 10

# Проверяем статус сервисов
echo "📊 Checking service status..."
docker-compose ps

# Проверяем доступность сервисов
echo "🔍 Checking service availability..."

# Проверяем Loki
if curl -s http://localhost:3100/ready > /dev/null; then
    echo "✅ Loki is ready"
else
    echo "❌ Loki is not ready"
fi

# Проверяем Prometheus
if curl -s http://localhost:9090/-/healthy > /dev/null; then
    echo "✅ Prometheus is ready"
else
    echo "❌ Prometheus is not ready"
fi

# Проверяем Grafana
if curl -s http://localhost:3000/api/health > /dev/null; then
    echo "✅ Grafana is ready"
else
    echo "❌ Grafana is not ready"
fi

echo ""
echo "🎉 Monitoring system is starting up!"
echo ""
echo "📊 Access URLs:"
echo "   Grafana:    http://localhost:3000 (admin/admin)"
echo "   Prometheus: http://localhost:9090"
echo "   Loki:       http://localhost:3100"
echo ""
echo "📈 Available Dashboards:"
echo "   - Smart Carwash - Logs Dashboard"
echo "   - Smart Carwash - System Metrics"
echo "   - Smart Carwash - Application Metrics"
echo ""
echo "📚 Documentation: ./monitoring/README.md"
echo ""
echo "🛑 To stop monitoring: docker-compose down"
echo "📋 To view logs: docker-compose logs -f"
echo ""
echo "⏳ Services are starting up. Please wait a few minutes for all services to be fully ready."

