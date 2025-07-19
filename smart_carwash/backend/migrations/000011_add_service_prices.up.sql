-- Добавляем базовые цены для услуг
INSERT INTO service_settings (id, service_type, setting_key, setting_value, created_at, updated_at) VALUES
-- Цены для мойки (wash)
(gen_random_uuid(), 'wash', 'price_per_minute', '1000', NOW(), NOW()),
(gen_random_uuid(), 'wash', 'chemistry_price_per_minute', '200', NOW(), NOW()),

-- Цены для обдува воздухом (air_dry)
(gen_random_uuid(), 'air_dry', 'price_per_minute', '600', NOW(), NOW()),
(gen_random_uuid(), 'air_dry', 'chemistry_price_per_minute', '100', NOW(), NOW()),

-- Цены для пылесоса (vacuum)
(gen_random_uuid(), 'vacuum', 'price_per_minute', '400', NOW(), NOW()),
(gen_random_uuid(), 'vacuum', 'chemistry_price_per_minute', '50', NOW(), NOW()); 