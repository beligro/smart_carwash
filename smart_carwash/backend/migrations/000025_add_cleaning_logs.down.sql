-- Удаление триггера
DROP TRIGGER IF EXISTS trigger_update_cleaning_logs_updated_at ON cleaning_logs;

-- Удаление функции
DROP FUNCTION IF EXISTS update_cleaning_logs_updated_at();

-- Удаление таблицы логов уборки
DROP TABLE IF EXISTS cleaning_logs;

-- Восстановление настройки времени уборки на 30 минут
UPDATE service_settings 
SET setting_value = '30' 
WHERE service_type = 'cleaner' AND setting_key = 'cleaning_timeout_minutes';

