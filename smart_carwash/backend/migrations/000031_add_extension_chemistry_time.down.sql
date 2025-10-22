-- Удалить поле для времени химии при продлении сессии
ALTER TABLE sessions DROP COLUMN extension_chemistry_time_minutes;