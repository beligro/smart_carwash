#!/bin/bash

# Простой скрипт для обновления SSL сертификатов с использованием Docker
# Автор: Smart Carwash System
# Версия: 1.0

set -e

# Конфигурация
DOMAIN="h2o-nsk.ru"
NGINX_SSL_DIR="/home/artem/smart_carwash/smart_carwash/nginx/ssl"
PROJECT_DIR="/home/artem/smart_carwash/smart_carwash"
LOG_FILE="/home/artem/smart_carwash/smart_carwash/logs/ssl_renew.log"

# Функция логирования
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Функция проверки срока действия сертификата
check_certificate_expiry() {
    local cert_file="$NGINX_SSL_DIR/fullchain.pem"
    
    if [ ! -f "$cert_file" ]; then
        log "ERROR: Сертификат не найден: $cert_file"
        return 1
    fi
    
    local expiry_date=$(openssl x509 -in "$cert_file" -noout -enddate | cut -d= -f2)
    local expiry_timestamp=$(date -d "$expiry_date" +%s)
    local current_timestamp=$(date +%s)
    local days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
    
    log "Сертификат истекает: $expiry_date (через $days_until_expiry дней)"
    
    if [ $days_until_expiry -lt 30 ]; then
        log "WARNING: Сертификат истекает менее чем через 30 дней!"
        return 0
    else
        log "INFO: Сертификат действителен более 30 дней"
        return 1
    fi
}

# Функция обновления сертификата
renew_certificate() {
    log "Начинаем обновление SSL сертификата для домена: $DOMAIN"
    
    # Останавливаем nginx
    log "Останавливаем nginx контейнер..."
    cd "$PROJECT_DIR"
    docker compose stop nginx || true
    
    # Создаем резервную копию
    log "Создаем резервную копию текущих сертификатов..."
    local backup_dir="$NGINX_SSL_DIR/backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    cp "$NGINX_SSL_DIR/fullchain.pem" "$backup_dir/" 2>/dev/null || true
    cp "$NGINX_SSL_DIR/privkey.pem" "$backup_dir/" 2>/dev/null || true
    
    # Создаем директории для certbot
    mkdir -p "$PROJECT_DIR/certbot/conf"
    mkdir -p "$PROJECT_DIR/certbot/www"
    
    # Запускаем certbot в Docker
    log "Запускаем certbot в Docker контейнере..."
    if docker run --rm \
        -p 80:80 \
        -v "$PROJECT_DIR/certbot/conf:/etc/letsencrypt" \
        -v "$PROJECT_DIR/certbot/www:/var/www/certbot" \
        certbot/certbot \
        certonly --standalone \
        --agree-tos \
        --no-eff-email \
        --email admin@$DOMAIN \
        --force-renewal \
        -d "$DOMAIN"; then
        
        log "Сертификат успешно обновлен!"
        
        # Копируем новые сертификаты
        log "Копируем новые сертификаты в nginx/ssl..."
        cp "$PROJECT_DIR/certbot/conf/live/$DOMAIN/fullchain.pem" "$NGINX_SSL_DIR/"
        cp "$PROJECT_DIR/certbot/conf/live/$DOMAIN/privkey.pem" "$NGINX_SSL_DIR/"
        
        # Устанавливаем права доступа
        chmod 644 "$NGINX_SSL_DIR/fullchain.pem"
        chmod 600 "$NGINX_SSL_DIR/privkey.pem"
        
        # Перезапускаем nginx
        log "Перезапускаем nginx контейнер..."
        docker compose up -d nginx
        
        # Проверяем статус
        sleep 5
        if docker compose ps nginx | grep -q "Up"; then
            log "SUCCESS: Nginx успешно перезапущен с новым сертификатом"
            return 0
        else
            log "ERROR: Nginx не запустился после обновления сертификата"
            return 1
        fi
    else
        log "ERROR: Не удалось обновить сертификат"
        
        # Восстанавливаем nginx
        log "Восстанавливаем nginx..."
        docker compose up -d nginx
        return 1
    fi
}

# Основная функция
main() {
    log "=== Запуск проверки SSL сертификата ==="
    
    # Проверяем срок действия сертификата
    if check_certificate_expiry; then
        log "Сертификат требует обновления"
        if renew_certificate; then
            log "Сертификат успешно обновлен"
        else
            log "ERROR: Не удалось обновить сертификат"
            exit 1
        fi
    else
        log "Сертификат действителен, обновление не требуется"
    fi
    
    log "=== Проверка SSL сертификата завершена ==="
}

# Обработка аргументов командной строки
case "${1:-}" in
    "force")
        log "Принудительное обновление сертификата"
        renew_certificate
        ;;
    "check")
        log "Проверка статуса сертификата"
        check_certificate_expiry
        ;;
    *)
        main
        ;;
esac
