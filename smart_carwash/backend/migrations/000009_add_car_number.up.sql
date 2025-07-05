-- Добавление поля car_number в таблицу users
ALTER TABLE users ADD COLUMN IF NOT EXISTS car_number VARCHAR(20);

-- Добавление поля car_number в таблицу sessions
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS car_number VARCHAR(20);

-- Создание индекса для быстрого поиска по номеру машины в сессиях
CREATE INDEX IF NOT EXISTS idx_sessions_car_number ON sessions(car_number); 