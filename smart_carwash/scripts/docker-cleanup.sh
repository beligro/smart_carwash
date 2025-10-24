#!/bin/bash

# Скрипт автоматической очистки Docker для Smart Carwash
# Запускается еженедельно через cron

echo "$(date): Начинаем очистку Docker..."

# Очистка неиспользуемых контейнеров, сетей, образов и кэша сборки
docker system prune -f

# Очистка неиспользуемых образов (более агрессивная)
docker image prune -a -f

# Очистка неиспользуемых сетей
docker network prune -f

# Проверка использования диска после очистки
DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
echo "$(date): Использование диска после очистки: ${DISK_USAGE}%"

# Если диск все еще заполнен более чем на 80%, дополнительная очистка
if [ "$DISK_USAGE" -gt 80 ]; then
    echo "$(date): Диск заполнен более чем на 80%, выполняем дополнительную очистку..."
    
    # Удаление всех остановленных контейнеров
    docker container prune -f
    
    # Удаление всех неиспользуемых образов
    docker image prune -a -f --filter "until=24h"
    
    # Очистка кэша сборки
    docker builder prune -a -f
fi

# Финальная проверка
FINAL_DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
echo "$(date): Очистка завершена. Финальное использование диска: ${FINAL_DISK_USAGE}%"

# Логирование в файл
echo "$(date): Docker cleanup completed. Disk usage: ${FINAL_DISK_USAGE}%" >> /var/log/docker-cleanup.log
