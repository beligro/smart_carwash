-- Изменение типа поля car_number с VARCHAR(20) на VARCHAR(36) во всех таблицах
-- Это позволит хранить более длинные номера автомобилей

-- Изменение типа поля car_number в таблице users
ALTER TABLE users ALTER COLUMN car_number TYPE VARCHAR(36);

-- Изменение типа поля car_number в таблице sessions  
ALTER TABLE sessions ALTER COLUMN car_number TYPE VARCHAR(36);

-- Изменение типа поля last_completed_session_car_number в таблице wash_boxes
ALTER TABLE wash_boxes ALTER COLUMN last_completed_session_car_number TYPE VARCHAR(36);
