#!/bin/bash

# Скрипт настройки cron jobs для Smart Carwash

echo "Настройка cron jobs для Smart Carwash..."

# Создаем cron job для ежедневной очистки Docker (каждый день в 2:00)
(crontab -l 2>/dev/null; echo "0 2 * * * /home/artem/smart_carwash/smart_carwash/scripts/docker-cleanup.sh") | crontab -

# Создаем cron job для ежедневной проверки диска (каждый день в 6:00)
(crontab -l 2>/dev/null; echo "0 6 * * * df -h | grep -E '9[0-9]%|100%' && echo 'WARNING: Disk usage is high!' | mail -s 'Disk Usage Alert' root") | crontab -

# Создаем cron job для еженедельной очистки старых логов (каждую субботу в 3:00)
(crontab -l 2>/dev/null; echo "0 3 * * 6 find /var/log -name '*.log.*' -mtime +7 -delete") | crontab -

# Создаем cron job для еженедельной очистки Docker логов (каждую субботу в 3:30)
(crontab -l 2>/dev/null; echo "30 3 * * 6 find /var/lib/docker/containers -name '*.log' -mtime +7 -delete") | crontab -

# Создаем cron job для ежедневной очистки systemd journal (каждый день в 4:00)
(crontab -l 2>/dev/null; echo "0 4 * * * /home/artem/smart_carwash/smart_carwash/scripts/cleanup-journal.sh") | crontab -

echo "Cron jobs настроены:"
echo "- Ежедневная очистка Docker: каждый день 2:00"
echo "- Ежедневная проверка диска: каждый день 6:00"
echo "- Еженедельная очистка логов: суббота 3:00"
echo "- Еженедельная очистка Docker логов: суббота 3:30"
echo "- Ежедневная очистка systemd journal: каждый день 4:00"

# Показываем текущие cron jobs
echo "Текущие cron jobs:"
crontab -l
