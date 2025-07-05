-- Удаление индекса
DROP INDEX IF EXISTS idx_sessions_car_number;

-- Удаление поля car_number из таблицы sessions
ALTER TABLE sessions DROP COLUMN IF EXISTS car_number;

-- Удаление поля car_number из таблицы users
ALTER TABLE users DROP COLUMN IF EXISTS car_number; 