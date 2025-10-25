#!/bin/bash

# Скрипт настройки systemd journal для предотвращения заполнения диска

echo "Настройка systemd journal для ограничения размера..."

# Создаем конфигурацию journal
sudo tee /etc/systemd/journald.conf.d/99-smart-carwash.conf > /dev/null <<EOF
[Journal]
# Ограничиваем размер journal
SystemMaxUse=100M
SystemMaxFileSize=10M
SystemMaxFiles=10

# Ограничиваем время хранения
MaxRetentionSec=7d

# Сжимаем старые файлы
Compress=yes
EOF

echo "Конфигурация journal создана:"
echo "- Максимальный размер: 100MB"
echo "- Максимальный размер файла: 10MB"
echo "- Максимальное количество файлов: 10"
echo "- Время хранения: 7 дней"
echo "- Сжатие: включено"

# Перезапускаем journald
sudo systemctl restart systemd-journald

echo "Journald перезапущен с новыми настройками"
