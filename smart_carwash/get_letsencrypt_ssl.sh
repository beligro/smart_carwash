#!/bin/bash

# Скрипт для получения SSL сертификата с помощью Let's Encrypt и сервиса nip.io

# Проверяем, установлен ли certbot
if ! command -v certbot &> /dev/null; then
    echo "certbot не установлен. Устанавливаем..."
    sudo apt-get update
    sudo apt-get install -y certbot
fi

# Получаем IP-адрес сервера из .env файла
SERVER_IP=$(grep SERVER_IP .env | cut -d '=' -f2)

# Проверяем, что IP-адрес получен
if [ -z "$SERVER_IP" ]; then
    echo "Не удалось получить IP-адрес сервера из .env файла."
    exit 1
fi

# Создаем доменное имя на основе IP-адреса с помощью сервиса nip.io
DOMAIN="${SERVER_IP}.nip.io"

echo "Получаем SSL сертификат для домена: $DOMAIN (IP: $SERVER_IP)"

# Создаем директорию для SSL сертификатов, если она не существует
mkdir -p nginx/ssl

# Останавливаем nginx, если он запущен, чтобы освободить 80 порт для certbot
echo "Останавливаем nginx для освобождения 80 порта..."
docker-compose stop nginx || true

# Получаем сертификат с помощью certbot и Let's Encrypt
echo "Запускаем certbot для получения сертификата..."
sudo certbot certonly --standalone --preferred-challenges http --agree-tos -d $DOMAIN

# Проверяем, что сертификат получен успешно
# Используем sudo для проверки наличия директории, так как она доступна только для root
if sudo test -d "/etc/letsencrypt/live/$DOMAIN"; then
    echo "Сертификат успешно получен!"
    
    # Копируем сертификаты в директорию nginx/ssl
    echo "Копируем сертификаты в директорию nginx/ssl..."
    sudo cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem nginx/ssl/
    sudo cp /etc/letsencrypt/live/$DOMAIN/privkey.pem nginx/ssl/
    
    # Устанавливаем правильные права доступа
    sudo chmod 644 nginx/ssl/fullchain.pem
    sudo chmod 600 nginx/ssl/privkey.pem
    sudo chown $(whoami):$(whoami) nginx/ssl/fullchain.pem nginx/ssl/privkey.pem
    
    # Обновляем .env файл, заменяя SERVER_IP на DOMAIN
    echo "Обновляем .env файл..."
    sed -i "s/SERVER_IP=$SERVER_IP/SERVER_IP=$DOMAIN/g" .env
    
    # Обновляем nginx.conf, заменяя IP-адрес на доменное имя
    echo "Обновляем nginx.conf..."
    sed -i "s/server_name $SERVER_IP;/server_name $DOMAIN;/g" nginx/nginx.conf
    
    # Обновляем URL в боте
    echo "Обновляем URL в боте..."
    sed -i "s/https:\/\/$SERVER_IP/https:\/\/$DOMAIN/g" backend/internal/telegram/bot.go
    
    # Обновляем URL в ApiService.js
    echo "Обновляем URL в ApiService.js..."
    sed -i "s/https:\/\/$SERVER_IP/https:\/\/$DOMAIN/g" frontend/src/services/ApiService.js
    
    echo "Готово! Теперь вы можете запустить проект с доверенным SSL сертификатом:"
    echo "make restart"
    
    echo "Ваше приложение будет доступно по адресу: https://$DOMAIN"
    echo "Все URL в проекте были обновлены для использования доменного имени вместо IP-адреса."
else
    echo "Не удалось получить сертификат или нет доступа к директории сертификатов."
    echo "Проверьте логи certbot: sudo cat /var/log/letsencrypt/letsencrypt.log"
fi
