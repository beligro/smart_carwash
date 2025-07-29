-- Добавление поля для хранения запрошенного времени продления
ALTER TABLE sessions ADD COLUMN requested_extension_time_minutes INT DEFAULT 0; 