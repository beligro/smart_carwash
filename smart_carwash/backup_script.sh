#!/bin/bash

# Оптимизированный скрипт бэкапа PostgreSQL
# Включает проверку целостности, ротацию и логирование

set -e

# Переменные окружения
export PGPASSWORD=${POSTGRES_PASSWORD}
BACKUP_DIR="/backups"
MAX_BACKUPS=7  # Хранить последние 7 бэкапов

# Функция логирования
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Функция создания бэкапа
create_backup() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="${BACKUP_DIR}/backup_${timestamp}.dump"
    
    log "Начинаем создание бэкапа: ${backup_file}"
    local start_time=$(date +%s)
    
    # Создание оптимизированного бэкапа
    pg_dump \
        -h postgres \
        -U ${POSTGRES_USER} \
        --format=custom \
        --no-owner \
        --no-privileges \
        --no-sync \
        --verbose \
        --file="${backup_file}" \
        ${POSTGRES_DB}
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log "Бэкап создан за ${duration} секунд"
    
    # Проверка целостности бэкапа
    log "Проверяем целостность бэкапа..."
    if pg_restore --list "${backup_file}" > /dev/null 2>&1; then
        log "Бэкап прошел проверку целостности"
    else
        log "ОШИБКА: Бэкап поврежден!"
        rm -f "${backup_file}"
        return 1
    fi
    
    # Сжатие бэкапа
    log "Сжимаем бэкап..."
    gzip "${backup_file}"
    local compressed_file="${backup_file}.gz"
    
    # Проверка размера файла
    local file_size=$(du -h "${compressed_file}" | cut -f1)
    log "Размер сжатого бэкапа: ${file_size}"
    
    return 0
}

# Функция ротации бэкапов
rotate_backups() {
    log "Выполняем ротацию бэкапов..."
    
    # Подсчет количества бэкапов
    local backup_count=$(ls -1 ${BACKUP_DIR}/backup_*.dump.gz 2>/dev/null | wc -l)
    
    if [ ${backup_count} -gt ${MAX_BACKUPS} ]; then
        local to_delete=$((backup_count - MAX_BACKUPS))
        log "Удаляем ${to_delete} старых бэкапов"
        
        # Удаляем самые старые файлы
        ls -1t ${BACKUP_DIR}/backup_*.dump.gz | tail -n ${to_delete} | xargs rm -f
    fi
    
    log "Ротация завершена. Осталось бэкапов: $(ls -1 ${BACKUP_DIR}/backup_*.dump.gz 2>/dev/null | wc -l)"
}

# Основной цикл
while true; do
    log "=== Начинаем процедуру бэкапа ==="
    
    if create_backup; then
        rotate_backups
        log "=== Бэкап успешно завершен ==="
    else
        log "=== ОШИБКА при создании бэкапа ==="
    fi
    
    log "Ожидаем 24 часа до следующего бэкапа..."
    sleep 86400
done
