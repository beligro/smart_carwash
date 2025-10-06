-- Удаление настройки времени уборки
DELETE FROM service_settings WHERE service_type = 'cleaner' AND setting_key = 'cleaning_timeout_minutes';

-- Удаление индексов
DROP INDEX IF EXISTS idx_washboxes_cleaning_started_at;
DROP INDEX IF EXISTS idx_washboxes_cleaning_reserved_by;
DROP INDEX IF EXISTS idx_cleaner_sessions_token;
DROP INDEX IF EXISTS idx_cleaner_sessions_cleaner_id;
DROP INDEX IF EXISTS idx_cleaners_is_active;
DROP INDEX IF EXISTS idx_cleaners_username;

-- Удаление внешнего ключа
ALTER TABLE wash_boxes DROP CONSTRAINT IF EXISTS fk_washboxes_cleaning_reserved_by;

-- Удаление полей из таблицы wash_boxes
ALTER TABLE wash_boxes DROP COLUMN IF EXISTS cleaning_started_at;
ALTER TABLE wash_boxes DROP COLUMN IF EXISTS cleaning_reserved_by;

-- Удаление таблиц
DROP TABLE IF EXISTS cleaner_sessions;
DROP TABLE IF EXISTS cleaners;
