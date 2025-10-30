#!/bin/bash

# Скрипт мониторинга производительности для выявления причин зависаний

LOG_FILE="/home/artem/smart_carwash/smart_carwash/logs/performance_monitor.log"

log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

check_docker_stats() {
    log_message "🐳 Docker статистика:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" | tee -a "$LOG_FILE"
}

check_database_connections() {
    log_message "🗄️ Соединения с БД:"
    docker exec carwash_postgres psql -U postgres -d carwash -c "SELECT COUNT(*) as total_connections, state, COUNT(*) * 100.0 / SUM(COUNT(*)) OVER() as percentage FROM pg_stat_activity GROUP BY state ORDER BY total_connections DESC;" | tee -a "$LOG_FILE"
}

check_database_locks() {
    log_message "🔒 Блокировки БД:"
    docker exec carwash_postgres psql -U postgres -d carwash -c "SELECT blocked_locks.pid AS blocked_pid, blocked_activity.usename AS blocked_user, blocking_locks.pid AS blocking_pid, blocking_activity.usename AS blocking_user, blocked_activity.query AS blocked_statement FROM pg_catalog.pg_locks blocked_locks JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid AND blocking_locks.pid != blocked_locks.pid JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid WHERE NOT blocked_locks.granted;" | tee -a "$LOG_FILE"
}

check_backend_logs() {
    log_message "📋 Последние ошибки backend:"
    docker logs carwash_backend --tail=10 2>&1 | grep -i "error\|fatal\|panic" | tail -5 | tee -a "$LOG_FILE"
}

check_system_load() {
    log_message "⚡ Системная нагрузка:"
    uptime | tee -a "$LOG_FILE"
    free -h | tee -a "$LOG_FILE"
    df -h / | tee -a "$LOG_FILE"
}

check_network_connections() {
    log_message "🌐 Сетевые соединения:"
    netstat -tuln | grep -E ":(80|443|8080|5432|9090|3000|3100|9100)" | tee -a "$LOG_FILE"
}

main() {
    log_message "🔍 Начинаем мониторинг производительности..."
    
    check_docker_stats
    check_database_connections
    check_database_locks
    check_backend_logs
    check_system_load
    check_network_connections
    
    log_message "✅ Мониторинг завершен"
}

main "$@"
