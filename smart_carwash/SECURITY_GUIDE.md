# 🛡️ Руководство по безопасности Smart Carwash

## Установленная защита от криптомайнеров

### 1. Файрвол (UFW)
- **Файл:** `setup_firewall.sh`
- **Функция:** Блокирует подозрительные порты и ограничивает доступ
- **Запуск:** `./setup_firewall.sh`

### 2. Мониторинг безопасности
- **Файл:** `security_monitor.sh`
- **Функция:** Автоматическая проверка на криптомайнеры каждые 5 минут
- **Логи:** `/var/log/security_monitor.log`

### 3. Защищенная конфигурация PostgreSQL
- **Файл:** `postgres_security.conf`
- **Функция:** Ограничения подключений, SSL, аудит
- **Применение:** Автоматически при запуске контейнера

### 4. Безопасные Docker контейнеры
- **Read-only файловая система**
- **No new privileges**
- **Ограниченный доступ к портам**
- **Tmpfs для временных файлов**

### 5. Автоматическая защита
- **Файл:** `auto_protection.sh`
- **Функция:** Устанавливает cron задачи и systemd таймеры
- **Запуск:** `./auto_protection.sh`

### 6. Экстренное реагирование
- **Файл:** `emergency_response.sh`
- **Функция:** Автоматическая очистка и восстановление
- **Запуск:** `./emergency_response.sh`

## Команды для управления

### Проверка статуса защиты
```bash
# Статус файрвола
sudo ufw status

# Статус мониторинга
sudo systemctl status security-monitor.timer

# Последние логи
tail -f /var/log/security_monitor.log
```

### Ручная проверка на криптомайнеры
```bash
# Проверка процессов
ps aux | grep -E "(kinsing|kdevtmpfsi|libsystem|xmrig)"

# Проверка Docker контейнеров
docker exec carwash_postgres ps aux | grep -E "(kinsing|kdevtmpfsi)"

# Проверка загрузки CPU
top -bn1 | head -5
```

### Экстренные действия
```bash
# Если обнаружен криптомайнер
./emergency_response.sh

# Остановка всех контейнеров
docker-compose down

# Пересоздание с нуля
docker-compose down -v
docker-compose up -d
```

## Рекомендации

1. **Регулярно обновляйте систему:**
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

2. **Мониторьте логи:**
   ```bash
   tail -f /var/log/security_monitor.log
   ```

3. **Проверяйте статус защиты:**
   ```bash
   sudo systemctl status security-monitor.timer
   ```

4. **Делайте резервные копии:**
   ```bash
   docker exec carwash_postgres pg_dump -U postgres carwash > backup_$(date +%Y%m%d_%H%M%S).sql
   ```

## Контакты для экстренных случаев
- **Логи:** `/var/log/security_monitor.log`
- **Статус:** `sudo systemctl status security-monitor.timer`
