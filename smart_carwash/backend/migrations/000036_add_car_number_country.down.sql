-- Удаление индексов для поиска по стране гос номера
DROP INDEX IF EXISTS idx_sessions_car_number_country;
DROP INDEX IF EXISTS idx_users_car_number_country;

-- Удаление поля country из таблиц sessions и users
ALTER TABLE sessions DROP COLUMN IF EXISTS car_number_country;
ALTER TABLE users DROP COLUMN IF EXISTS car_number_country;
