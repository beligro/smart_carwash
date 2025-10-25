#!/bin/bash

# Скрипт очистки systemd journal без sudo (только для пользователя)

echo "$(date): Начинаем очистку systemd journal (без sudo)..."

# Очистка journal пользователя (если есть)
journalctl --user --vacuum-time=1d 2>/dev/null || echo "Нет пользовательских journal логов"

# Проверка размера journal (только чтение)
JOURNAL_SIZE=$(journalctl --disk-usage 2>/dev/null | grep -o '[0-9.]*M' || echo "недоступно")
echo "$(date): Размер journal: $JOURNAL_SIZE"

# Проверка использования диска
DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
echo "$(date): Использование диска: ${DISK_USAGE}%"

# Логирование в домашнюю директорию
echo "$(date): Journal cleanup attempt completed. Size: $JOURNAL_SIZE, Disk usage: ${DISK_USAGE}%" >> /home/artem/journal-cleanup.log

echo "Для полной очистки journal нужны права sudo. Обратитесь к администратору."
