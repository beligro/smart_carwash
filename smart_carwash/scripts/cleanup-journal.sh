#!/bin/bash

# Скрипт очистки systemd journal для предотвращения заполнения диска

echo "$(date): Начинаем очистку systemd journal..."

# Очистка journal старше 1 дня
sudo journalctl --vacuum-time=1d

# Ограничение размера journal до 100MB
sudo journalctl --vacuum-size=100M

# Проверка размера после очистки
JOURNAL_SIZE=$(sudo journalctl --disk-usage | grep -o '[0-9.]*M')
echo "$(date): Размер journal после очистки: $JOURNAL_SIZE"

# Проверка использования диска
DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
echo "$(date): Использование диска: ${DISK_USAGE}%"

# Логирование
echo "$(date): Journal cleanup completed. Size: $JOURNAL_SIZE, Disk usage: ${DISK_USAGE}%" >> /var/log/journal-cleanup.log
