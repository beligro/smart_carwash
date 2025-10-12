#!/bin/bash

# Скрипт для автоматического обновления SSL сертификатов Let's Encrypt
# Автор: Smart Carwash System
# Версия: 1.0

set -e

# Конфигурация
DOMAIN="h2o-nsk.ru"
NGINX_SSL_DIR="/home/artem/smart_carwash/smart_carwash/nginx/ssl"
PROJECT_DIR="/home/artem/smart_carwash/smart_carwash"
LOG_FILE="/home/artem/smart_carwash/smart_carwash/logs/ssl_auto_renew.log"
TELEGRAM_BOT_TOKEN="8100164537:AAG3sTeSDrDSQaPyifCCAjsGzmTA7cqbQJA"
TELEGRAM_CHAT_ID="" # Будет заполнено при первом запуске

# Функция логирования
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Функция отправки уведомления в Telegram
send_telegram_notification() {
    local message="$1"
    if [ -n "$TELEGRAM_CHAT_ID" ]; then
        curl -s -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
            -d chat_id="$TELEGRAM_CHAT_ID" \
            -d text="$message" \
            -d parse_mode="HTML" > /dev/null 2>&1 || true
    fi
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
    
    # Останавливаем nginx для освобождения 80 порта
    log "Останавливаем nginx контейнер..."
    cd "$PROJECT_DIR"
    docker compose stop nginx || true
    
    # Создаем резервную копию текущих сертификатов
    log "Создаем резервную копию текущих сертификатов..."
    local backup_dir="$NGINX_SSL_DIR/backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    cp "$NGINX_SSL_DIR/fullchain.pem" "$backup_dir/" 2>/dev/null || true
    cp "$NGINX_SSL_DIR/privkey.pem" "$backup_dir/" 2>/dev/null || true
    
    # Обновляем сертификат
    log "Запускаем certbot для обновления сертификата..."
    if sudo certbot renew --standalone --preferred-challenges http --agree-tos -d "$DOMAIN" --force-renewal; then
        log "Сертификат успешно обновлен!"
        
        # Копируем новые сертификаты
        log "Копируем новые сертификаты в nginx/ssl..."
        sudo cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" "$NGINX_SSL_DIR/"
        sudo cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" "$NGINX_SSL_DIR/"
        
        # Устанавливаем правильные права доступа
        sudo chmod 644 "$NGINX_SSL_DIR/fullchain.pem"
        sudo chmod 600 "$NGINX_SSL_DIR/privkey.pem"
        sudo chown $(whoami):$(whoami) "$NGINX_SSL_DIR/fullchain.pem" "$NGINX_SSL_DIR/privkey.pem"
        
        # Перезапускаем nginx
        log "Перезапускаем nginx контейнер..."
        docker compose up -d nginx
        
        # Проверяем, что nginx запустился успешно
        sleep 5
        if docker compose ps nginx | grep -q "Up"; then
            log "SUCCESS: Nginx успешно перезапущен с новым сертификатом"
            send_telegram_notification "✅ SSL сертификат для $DOMAIN успешно обновлен и nginx перезапущен"
            return 0
        else
            log "ERROR: Nginx не запустился после обновления сертификата"
            send_telegram_notification "❌ Ошибка: Nginx не запустился после обновления SSL сертификата для $DOMAIN"
            return 1
        fi
    else
        log "ERROR: Не удалось обновить сертификат"
        send_telegram_notification "❌ Ошибка: Не удалось обновить SSL сертификат для $DOMAIN"
        
        # Восстанавливаем nginx
        log "Восстанавливаем nginx..."
        docker compose up -d nginx
        return 1
    fi
}

# Функция проверки статуса nginx
check_nginx_status() {
    cd "$PROJECT_DIR"
    if docker compose ps nginx | grep -q "Up"; then
        log "INFO: Nginx работает нормально"
        return 0
    else
        log "ERROR: Nginx не работает"
        return 1
    fi
}

# Функция тестирования SSL соединения
test_ssl_connection() {
    local test_url="https://$DOMAIN"
    log "Тестируем SSL соединение: $test_url"
    
    if curl -s --connect-timeout 10 --max-time 30 "$test_url" > /dev/null 2>&1; then
        log "SUCCESS: SSL соединение работает нормально"
        return 0
    else
        log "ERROR: SSL соединение не работает"
        return 1
    fi
}

# Основная функция
main() {
    log "=== Запуск проверки SSL сертификата ==="
    
    # Проверяем, что мы в правильной директории
    if [ ! -f "$PROJECT_DIR/docker-compose.yml" ]; then
        log "ERROR: docker-compose.yml не найден в $PROJECT_DIR"
        exit 1
    fi
    
    # Проверяем статус nginx
    if ! check_nginx_status; then
        log "WARNING: Nginx не работает, пытаемся запустить..."
        cd "$PROJECT_DIR"
        docker compose up -d nginx
        sleep 10
        if ! check_nginx_status; then
            log "ERROR: Не удалось запустить nginx"
            send_telegram_notification "❌ Критическая ошибка: Nginx не работает на $DOMAIN"
            exit 1
        fi
    fi
    
    # Проверяем срок действия сертификата
    if check_certificate_expiry; then
        log "Сертификат требует обновления"
        if renew_certificate; then
            log "Сертификат успешно обновлен"
            # Тестируем новое соединение
            sleep 10
            if test_ssl_connection; then
                log "SUCCESS: Все проверки пройдены успешно"
            else
                log "WARNING: SSL соединение не работает после обновления"
                send_telegram_notification "⚠️ Предупреждение: SSL соединение не работает после обновления для $DOMAIN"
            fi
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
    "test")
        log "Тестирование SSL соединения"
        test_ssl_connection
        ;;
    *)
        main
        ;;
esac
