#!/bin/bash

echo "🛡️ Настройка файрвола для защиты от криптомайнеров..."

# Включаем файрвол
sudo ufw --force enable

# Базовые правила
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Разрешаем SSH (важно!)
sudo ufw allow ssh

# Разрешаем HTTP и HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Разрешаем PostgreSQL только локально
sudo ufw allow from 127.0.0.1 to any port 5432
sudo ufw allow from 172.16.0.0/12 to any port 5432  # Docker networks

# Блокируем подозрительные порты
sudo ufw deny 4444/tcp  # Часто используется майнерами
sudo ufw deny 5555/tcp
sudo ufw deny 6666/tcp
sudo ufw deny 7777/tcp
sudo ufw deny 8888/tcp
sudo ufw deny 9999/tcp

# Блокируем известные порты майнеров
sudo ufw deny 8080/tcp  # Если не нужен
sudo ufw deny 8081/tcp
sudo ufw deny 8082/tcp

echo "✅ Файрвол настроен!"
sudo ufw status verbose
