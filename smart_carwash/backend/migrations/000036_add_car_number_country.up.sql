-- Добавление поля country для хранения страны гос номера
ALTER TABLE users ADD COLUMN IF NOT EXISTS car_number_country VARCHAR(3) DEFAULT 'RUS';
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS car_number_country VARCHAR(3) DEFAULT 'RUS';

-- Создание индекса для быстрого поиска по стране гос номера
CREATE INDEX IF NOT EXISTS idx_users_car_number_country ON users(car_number_country);
CREATE INDEX IF NOT EXISTS idx_sessions_car_number_country ON sessions(car_number_country);

-- Обновляем существующие записи - считаем их российскими
UPDATE users SET car_number_country = 'RUS' WHERE car_number_country IS NULL;
UPDATE sessions SET car_number_country = 'RUS' WHERE car_number_country IS NULL;
