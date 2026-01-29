#!/bin/sh
# Копирование SSL сертификатов при старте nginx
cp /etc/letsencrypt/live/h2o-nsk.ru/fullchain.pem /etc/nginx/ssl/fullchain.pem 2>/dev/null || echo "Certbot certs not found, using existing"
cp /etc/letsencrypt/live/h2o-nsk.ru/privkey.pem /etc/nginx/ssl/privkey.pem 2>/dev/null || echo "Certbot certs not found, using existing"
exec nginx -g 'daemon off;'
