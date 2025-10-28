-- Откат изменения типа поля car_number с VARCHAR(36) обратно на VARCHAR(20)
-- ВНИМАНИЕ: При откате могут быть потеряны данные, если есть записи с номерами длиннее 20 символов

-- Откат типа поля car_number в таблице users
ALTER TABLE users ALTER COLUMN car_number TYPE VARCHAR(20);

-- Откат типа поля car_number в таблице sessions
ALTER TABLE sessions ALTER COLUMN car_number TYPE VARCHAR(20);

-- Откат типа поля last_completed_session_car_number в таблице wash_boxes
ALTER TABLE wash_boxes ALTER COLUMN last_completed_session_car_number TYPE VARCHAR(20);
