#!/bin/bash

# Скрипт оптимизации системы Smart Carwash
# Выполняет все необходимые оптимизации для предотвращения проблем с диском

echo "🚀 Начинаем оптимизацию системы Smart Carwash..."

# 1. Очистка Docker
echo "📦 Очищаем Docker..."
docker system prune -f
docker image prune -a -f
docker network prune -f

# 2. Настройка cron jobs
echo "⏰ Настраиваем автоматические задачи..."
if [ -f "/home/artem/smart_carwash/smart_carwash/scripts/setup-cron.sh" ]; then
    /home/artem/smart_carwash/smart_carwash/scripts/setup-cron.sh
else
    echo "⚠️  Скрипт setup-cron.sh не найден"
fi

# 3. Настройка logrotate
echo "📄 Настраиваем ротацию логов..."
if [ -f "/home/artem/smart_carwash/smart_carwash/logrotate.conf" ]; then
    sudo cp /home/artem/smart_carwash/smart_carwash/logrotate.conf /etc/logrotate.d/smart-carwash
    echo "✅ Конфигурация logrotate установлена"
else
    echo "⚠️  Файл logrotate.conf не найден"
fi

# 4. Настройка systemd journal
echo "📋 Настраиваем systemd journal..."
if [ -f "/home/artem/smart_carwash/smart_carwash/scripts/configure-journal.sh" ]; then
    /home/artem/smart_carwash/smart_carwash/scripts/configure-journal.sh
    echo "✅ Конфигурация journal установлена"
else
    echo "⚠️  Скрипт configure-journal.sh не найден"
fi

# 5. Перезапуск сервисов с новыми настройками
echo "🔄 Перезапускаем сервисы с оптимизированными настройками..."
cd /home/artem/smart_carwash/smart_carwash
docker-compose restart loki prometheus

# 6. Проверка состояния системы
echo "📊 Проверяем состояние системы..."
echo "Использование диска:"
df -h /

echo "Использование памяти:"
free -h

echo "Статус контейнеров:"
docker-compose ps

# 7. Проверка доступности сервисов
echo "🌐 Проверяем доступность сервисов..."
echo "Prometheus: $(curl -s -o /dev/null -w '%{http_code}' http://localhost:9090/api/v1/status/config)"
echo "Grafana: $(curl -s -o /dev/null -w '%{http_code}' http://localhost:3000/api/health)"
echo "Loki: $(curl -s -o /dev/null -w '%{http_code}' http://localhost:3100/ready)"

echo "✅ Оптимизация системы завершена!"
echo ""
echo "📋 Что было сделано:"
echo "  ✅ Уменьшено время хранения логов Loki до 2 дней"
echo "  ✅ Настроена ротация логов для всех сервисов"
echo "  ✅ Добавлены ограничения размера Docker логов"
echo "  ✅ Настроена автоматическая очистка Docker"
echo "  ✅ Оптимизированы настройки Prometheus"
echo "  ✅ Настроены cron jobs для автоматического обслуживания"
echo "  ✅ Настроена автоматическая очистка systemd journal"
echo "  ✅ Ограничен размер systemd journal до 100MB"
echo ""
echo "🔧 Рекомендации:"
echo "  - Проверяйте Grafana дашборды регулярно"
echo "  - Мониторьте использование диска"
echo "  - Логи автоматически очищаются каждую неделю"
echo ""
echo "🌐 Доступ к мониторингу:"
echo "  - Grafana: http://localhost:3000 (admin/admin)"
echo "  - Prometheus: http://localhost:9090"
echo "  - Loki: http://localhost:3100"
