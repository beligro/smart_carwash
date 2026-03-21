#!/bin/bash

# Скрипт для проверки статуса SSL сертификата
# Автор: Smart Carwash System
# Версия: 1.0

set -e

# Конфигурация
DOMAIN="79.141.66.210.nip.io"
NGINX_SSL_DIR="/home/artem/smart_carwash/smart_carwash/nginx/ssl"
PROJECT_DIR="/home/artem/smart_carwash/smart_carwash"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция вывода с цветом
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Функция проверки срока действия сертификата
check_certificate_expiry() {
    local cert_file="$NGINX_SSL_DIR/fullchain.pem"
    
    if [ ! -f "$cert_file" ]; then
        print_status $RED "❌ Сертификат не найден: $cert_file"
        return 1
    fi
    
    local expiry_date=$(openssl x509 -in "$cert_file" -noout -enddate | cut -d= -f2)
    local expiry_timestamp=$(date -d "$expiry_date" +%s)
    local current_timestamp=$(date +%s)
    local days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
    
    print_status $BLUE "📅 Сертификат истекает: $expiry_date"
    
    if [ $days_until_expiry -lt 0 ]; then
        print_status $RED "❌ КРИТИЧНО: Сертификат истек $((days_until_expiry * -1)) дней назад!"
        return 2
    elif [ $days_until_expiry -lt 7 ]; then
        print_status $RED "⚠️  КРИТИЧНО: Сертификат истекает через $days_until_expiry дней!"
        return 1
    elif [ $days_until_expiry -lt 30 ]; then
        print_status $YELLOW "⚠️  ПРЕДУПРЕЖДЕНИЕ: Сертификат истекает через $days_until_expiry дней"
        return 1
    else
        print_status $GREEN "✅ Сертификат действителен еще $days_until_expiry дней"
        return 0
    fi
}

# Функция проверки статуса nginx
check_nginx_status() {
    cd "$PROJECT_DIR"
    if docker compose ps nginx | grep -q "Up"; then
        print_status $GREEN "✅ Nginx работает нормально"
        return 0
    else
        print_status $RED "❌ Nginx не работает"
        return 1
    fi
}

# Функция тестирования SSL соединения
test_ssl_connection() {
    local test_url="https://$DOMAIN"
    print_status $BLUE "🔍 Тестируем SSL соединение: $test_url"
    
    if curl -s --connect-timeout 10 --max-time 30 "$test_url" > /dev/null 2>&1; then
        print_status $GREEN "✅ SSL соединение работает нормально"
        return 0
    else
        print_status $RED "❌ SSL соединение не работает"
        return 1
    fi
}

# Функция проверки деталей сертификата
show_certificate_details() {
    local cert_file="$NGINX_SSL_DIR/fullchain.pem"
    
    if [ ! -f "$cert_file" ]; then
        print_status $RED "❌ Сертификат не найден: $cert_file"
        return 1
    fi
    
    print_status $BLUE "📋 Детали сертификата:"
    echo "----------------------------------------"
    openssl x509 -in "$cert_file" -text -noout | grep -E "(Subject:|Issuer:|Not Before|Not After|DNS:)"
    echo "----------------------------------------"
}

# Функция проверки цепочки сертификатов
check_certificate_chain() {
    local cert_file="$NGINX_SSL_DIR/fullchain.pem"
    
    if [ ! -f "$cert_file" ]; then
        print_status $RED "❌ Сертификат не найден: $cert_file"
        return 1
    fi
    
    print_status $BLUE "🔗 Проверка цепочки сертификатов:"
    openssl verify -CAfile "$cert_file" "$cert_file" 2>/dev/null && print_status $GREEN "✅ Цепочка сертификатов валидна" || print_status $RED "❌ Проблема с цепочкой сертификатов"
}

# Основная функция
main() {
    print_status $BLUE "🔐 === Проверка SSL сертификата для $DOMAIN ==="
    echo
    
    # Проверяем статус nginx
    print_status $BLUE "1. Проверка статуса Nginx..."
    check_nginx_status
    echo
    
    # Проверяем срок действия сертификата
    print_status $BLUE "2. Проверка срока действия сертификата..."
    local cert_status=$?
    check_certificate_expiry
    cert_status=$?
    echo
    
    # Показываем детали сертификата
    print_status $BLUE "3. Детали сертификата..."
    show_certificate_details
    echo
    
    # Проверяем цепочку сертификатов
    print_status $BLUE "4. Проверка цепочки сертификатов..."
    check_certificate_chain
    echo
    
    # Тестируем SSL соединение
    print_status $BLUE "5. Тестирование SSL соединения..."
    test_ssl_connection
    local ssl_status=$?
    echo
    
    # Итоговый статус
    print_status $BLUE "📊 === Итоговый статус ==="
    if [ $cert_status -eq 0 ] && [ $ssl_status -eq 0 ]; then
        print_status $GREEN "✅ Все проверки пройдены успешно"
        exit 0
    elif [ $cert_status -eq 2 ]; then
        print_status $RED "❌ КРИТИЧНО: Сертификат истек!"
        exit 2
    else
        print_status $YELLOW "⚠️  Обнаружены проблемы, требуется внимание"
        exit 1
    fi
}

# Обработка аргументов командной строки
case "${1:-}" in
    "expiry")
        check_certificate_expiry
        ;;
    "nginx")
        check_nginx_status
        ;;
    "ssl")
        test_ssl_connection
        ;;
    "details")
        show_certificate_details
        ;;
    "chain")
        check_certificate_chain
        ;;
    *)
        main
        ;;
esac
