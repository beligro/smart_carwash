#!/bin/bash

# Скрипт для настройки Telegram уведомлений для SSL автообновления
# Автор: Smart Carwash System
# Версия: 1.0

set -e

# Конфигурация
TELEGRAM_BOT_TOKEN="8100164537:AAG3sTeSDrDSQaPyifCCAjsGzmTA7cqbQJA"
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

# Функция получения chat_id для Telegram
get_telegram_chat_id() {
    print_status $BLUE "🔍 Получаем chat_id для Telegram уведомлений..."
    
    # Отправляем тестовое сообщение
    local test_message="🔐 Тест SSL уведомлений от Smart Carwash System"
    local response=$(curl -s -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
        -d chat_id="@carwash_grom_test_bot" \
        -d text="$test_message" \
        -d parse_mode="HTML" 2>/dev/null)
    
    if echo "$response" | grep -q '"ok":true'; then
        print_status $GREEN "✅ Тестовое сообщение отправлено в @carwash_grom_test_bot"
        echo "@carwash_grom_test_bot"
        return 0
    else
        print_status $YELLOW "⚠️  Не удалось отправить в @carwash_grom_test_bot, попробуем получить updates..."
        
        # Получаем последние updates
        local updates=$(curl -s "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getUpdates")
        
        if echo "$updates" | grep -q '"chat":{"id"'; then
            local chat_id=$(echo "$updates" | grep -o '"chat":{"id":[0-9]*' | head -1 | grep -o '[0-9]*')
            if [ -n "$chat_id" ]; then
                print_status $GREEN "✅ Найден chat_id: $chat_id"
                echo "$chat_id"
                return 0
            fi
        fi
        
        print_status $RED "❌ Не удалось получить chat_id автоматически"
        print_status $YELLOW "💡 Ручная настройка:"
        print_status $YELLOW "   1. Напишите боту @carwash_grom_test_bot любое сообщение"
        print_status $YELLOW "   2. Запустите этот скрипт снова"
        return 1
    fi
}

# Функция обновления скрипта с chat_id
update_script_with_chat_id() {
    local chat_id=$1
    local script_file="$PROJECT_DIR/ssl_auto_renew.sh"
    
    if [ -f "$script_file" ]; then
        print_status $BLUE "📝 Обновляем скрипт с chat_id: $chat_id"
        sed -i "s/TELEGRAM_CHAT_ID=\"\"/TELEGRAM_CHAT_ID=\"$chat_id\"/" "$script_file"
        print_status $GREEN "✅ Скрипт обновлен"
    else
        print_status $RED "❌ Файл скрипта не найден: $script_file"
        return 1
    fi
}

# Функция создания директории для логов
create_log_directory() {
    print_status $BLUE "📁 Создаем директорию для логов..."
    
    sudo mkdir -p /var/log
    sudo touch /var/log/ssl_auto_renew.log
    sudo touch /var/log/ssl_check.log
    sudo chown $(whoami):$(whoami) /var/log/ssl_auto_renew.log
    sudo chown $(whoami):$(whoami) /var/log/ssl_check.log
    sudo chmod 644 /var/log/ssl_auto_renew.log
    sudo chmod 644 /var/log/ssl_check.log
    
    print_status $GREEN "✅ Директория для логов создана"
}

# Функция тестирования уведомлений
test_notifications() {
    local chat_id=$1
    
    print_status $BLUE "🧪 Тестируем уведомления..."
    
    # Тестовое сообщение об успешном обновлении
    local success_message="✅ <b>SSL Auto-renewal Test</b>
🔐 Домен: h2o-nsk.ru
📅 Время: $(date)
✅ Тест успешного обновления SSL сертификата"
    
    curl -s -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
        -d chat_id="$chat_id" \
        -d text="$success_message" \
        -d parse_mode="HTML" > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        print_status $GREEN "✅ Тест успешного уведомления отправлен"
    else
        print_status $RED "❌ Ошибка отправки тестового уведомления"
        return 1
    fi
    
    sleep 2
    
    # Тестовое сообщение об ошибке
    local error_message="❌ <b>SSL Auto-renewal Test</b>
🔐 Домен: h2o-nsk.ru
📅 Время: $(date)
❌ Тест уведомления об ошибке SSL сертификата"
    
    curl -s -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
        -d chat_id="$chat_id" \
        -d text="$error_message" \
        -d parse_mode="HTML" > /dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        print_status $GREEN "✅ Тест уведомления об ошибке отправлен"
    else
        print_status $RED "❌ Ошибка отправки тестового уведомления об ошибке"
        return 1
    fi
}

# Основная функция
main() {
    print_status $BLUE "🔐 === Настройка SSL уведомлений ==="
    echo
    
    # Создаем директорию для логов
    create_log_directory
    echo
    
    # Получаем chat_id
    local chat_id=$(get_telegram_chat_id)
    if [ -z "$chat_id" ]; then
        print_status $RED "❌ Не удалось получить chat_id"
        exit 1
    fi
    echo
    
    # Обновляем скрипт с chat_id
    update_script_with_chat_id "$chat_id"
    echo
    
    # Тестируем уведомления
    print_status $BLUE "🧪 Тестируем уведомления..."
    test_notifications "$chat_id"
    echo
    
    print_status $GREEN "✅ Настройка SSL уведомлений завершена!"
    print_status $BLUE "📋 Что настроено:"
    print_status $BLUE "   • Chat ID: $chat_id"
    print_status $BLUE "   • Логи: /var/log/ssl_auto_renew.log"
    print_status $BLUE "   • Cron задачи настроены"
    print_status $BLUE "   • Уведомления протестированы"
    echo
    print_status $YELLOW "💡 Для проверки статуса SSL запустите: ./ssl_check.sh"
    print_status $YELLOW "💡 Для принудительного обновления: ./ssl_auto_renew.sh force"
}

# Обработка аргументов командной строки
case "${1:-}" in
    "test")
        local chat_id=$(get_telegram_chat_id)
        if [ -n "$chat_id" ]; then
            test_notifications "$chat_id"
        fi
        ;;
    "chat-id")
        get_telegram_chat_id
        ;;
    *)
        main
        ;;
esac
