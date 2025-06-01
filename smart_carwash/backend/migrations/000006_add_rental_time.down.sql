-- Удаление времени аренды из таблицы sessions
ALTER TABLE sessions DROP COLUMN IF EXISTS rental_time_minutes;

-- Удаление таблицы настроек
DROP TABLE IF EXISTS service_settings;
