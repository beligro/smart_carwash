-- Удаление поля для хранения запрошенного времени продления
ALTER TABLE sessions DROP COLUMN requested_extension_time_minutes; 