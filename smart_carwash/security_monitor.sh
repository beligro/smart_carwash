#!/bin/bash

# Создаем директорию для логов если её нет
mkdir -p /home/artem/smart_carwash/smart_carwash/logs
LOG_FILE="/home/artem/smart_carwash/smart_carwash/logs/security_monitor.log"
ALERT_EMAIL="admin@h2o-nsk.ru"  # Замените на ваш email

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

check_cryptominers() {
    log_message "🔍 Проверка на криптомайнеры..."
    
    # Проверяем процессы
    SUSPICIOUS_PROCESSES=$(ps aux | grep -E "(kinsing|kdevtmpfsi|libsystem|xmrig|cpuminer|ccminer|minerd)" | grep -v grep)
    
    if [ ! -z "$SUSPICIOUS_PROCESSES" ]; then
        log_message "🚨 ОБНАРУЖЕНЫ ПОДОЗРИТЕЛЬНЫЕ ПРОЦЕССЫ:"
        echo "$SUSPICIOUS_PROCESSES" >> "$LOG_FILE"
        # Отправляем алерт (если mail установлен)
        echo "Обнаружены подозрительные процессы на сервере!" | mail -s "SECURITY ALERT" "$ALERT_EMAIL" 2>/dev/null || true
        return 1
    fi
    
    # Проверяем высокую загрузку CPU (без bc, используем awk)
    CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1 | cut -d'.' -f1)
    if [ "$CPU_USAGE" -gt 80 ] 2>/dev/null; then
        log_message "⚠️ Высокая загрузка CPU: ${CPU_USAGE}%"
    fi
    
    log_message "✅ Проверка криптомайнеров завершена"
    return 0
}

check_docker_containers() {
    log_message "🐳 Проверка Docker контейнеров..."
    
    # Проверяем все контейнеры на подозрительные процессы
    for container in $(docker ps --format "{{.Names}}"); do
        SUSPICIOUS=$(docker exec "$container" ps aux 2>/dev/null | grep -E "(kinsing|kdevtmpfsi|libsystem|xmrig|cpuminer|ccminer|minerd)" | grep -v grep)
        
        if [ ! -z "$SUSPICIOUS" ]; then
            log_message "🚨 ПОДОЗРИТЕЛЬНЫЕ ПРОЦЕССЫ В КОНТЕЙНЕРЕ $container:"
            echo "$SUSPICIOUS" >> "$LOG_FILE"
            # Останавливаем контейнер
            docker stop "$container" 2>/dev/null || true
            log_message "🛑 Контейнер $container остановлен"
        fi
    done
    
    log_message "✅ Проверка контейнеров завершена"
}

check_network_connections() {
    log_message "🌐 Проверка сетевых подключений..."
    
    # Проверяем подозрительные подключения
    SUSPICIOUS_CONNECTIONS=$(netstat -tulpn 2>/dev/null | grep -E ":(4444|5555|6666|7777|8888|9999|8080|8081|8082)" | grep -v "127.0.0.1")
    
    if [ ! -z "$SUSPICIOUS_CONNECTIONS" ]; then
        log_message "🚨 ПОДОЗРИТЕЛЬНЫЕ СЕТЕВЫЕ ПОДКЛЮЧЕНИЯ:"
        echo "$SUSPICIOUS_CONNECTIONS" >> "$LOG_FILE"
    fi
    
    log_message "✅ Проверка сети завершена"
}

check_file_integrity() {
    log_message "📁 Проверка целостности файлов..."
    
    # Проверяем подозрительные файлы в /tmp
    SUSPICIOUS_FILES=$(find /tmp -name "*kinsing*" -o -name "*kdevtmpfsi*" -o -name "*libsystem*" -o -name "*xmrig*" 2>/dev/null)
    
    if [ ! -z "$SUSPICIOUS_FILES" ]; then
        log_message "🚨 ПОДОЗРИТЕЛЬНЫЕ ФАЙЛЫ НАЙДЕНЫ:"
        echo "$SUSPICIOUS_FILES" >> "$LOG_FILE"
        # Удаляем файлы
        echo "$SUSPICIOUS_FILES" | xargs rm -f 2>/dev/null || true
        log_message "🗑️ Подозрительные файлы удалены"
    fi
    
    log_message "✅ Проверка файлов завершена"
}

# Основная функция
main() {
    log_message "🛡️ Запуск мониторинга безопасности..."
    
    check_cryptominers
    check_docker_containers
    check_network_connections
    check_file_integrity
    
    log_message "✅ Мониторинг безопасности завершен"
}

# Запуск
main "$@"
