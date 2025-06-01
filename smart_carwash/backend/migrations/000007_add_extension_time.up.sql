-- Добавление времени продления в таблицу sessions
ALTER TABLE sessions ADD COLUMN extension_time_minutes INT DEFAULT 0;
