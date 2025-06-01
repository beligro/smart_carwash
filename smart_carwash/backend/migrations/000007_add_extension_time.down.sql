-- Удаление времени продления из таблицы sessions
ALTER TABLE sessions DROP COLUMN IF EXISTS extension_time_minutes;
