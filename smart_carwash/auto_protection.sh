#!/bin/bash

echo "🛡️ Установка автоматической защиты от криптомайнеров..."

# Создаем директорию для логов
mkdir -p /home/artem/smart_carwash/smart_carwash/logs

# Делаем скрипты исполняемыми
chmod +x security_monitor.sh
chmod +x setup_firewall.sh

# Создаем cron задачу для мониторинга каждые 5 минут
(crontab -l 2>/dev/null; echo "*/5 * * * * /home/artem/smart_carwash/smart_carwash/security_monitor.sh") | crontab -

echo "✅ Автоматическая защита настроена!"
echo "📊 Для запуска мониторинга выполните:"
echo "   ./security_monitor.sh"
echo ""
echo "📋 Логи будут сохраняться в:"
echo "   /home/artem/smart_carwash/smart_carwash/logs/security_monitor.log"
echo ""
echo "🔄 Для автоматического запуска каждые 5 минут добавьте в crontab:"
echo "   crontab -e"
echo "   */5 * * * * /home/artem/smart_carwash/smart_carwash/security_monitor.sh"
