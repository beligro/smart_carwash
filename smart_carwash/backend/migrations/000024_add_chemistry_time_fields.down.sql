-- Удалить настройки доступного времени химии
DELETE FROM service_settings WHERE setting_key = 'available_chemistry_times';

-- Удалить индекс
DROP INDEX IF EXISTS idx_sessions_chemistry_active;

-- Удалить поля времени химии из таблицы sessions
ALTER TABLE sessions DROP COLUMN IF EXISTS chemistry_ended_at;
ALTER TABLE sessions DROP COLUMN IF EXISTS chemistry_started_at;
ALTER TABLE sessions DROP COLUMN IF EXISTS chemistry_time_minutes;


