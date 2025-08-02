-- Добавление настроек цен для каждого типа услуги
INSERT INTO service_settings (service_type, setting_key, setting_value)
VALUES 
    ('wash', 'price_per_minute', '3000'), -- 30 рублей за минуту
    ('wash', 'price_per_minute_with_chemistry', '5000'), -- 50 рублей за минуту с химией
    ('air_dry', 'price_per_minute', '2000'), -- 20 рублей за минуту
    ('air_dry', 'price_per_minute_with_chemistry', '2000'), -- 20 рублей за минуту (без химии)
    ('vacuum', 'price_per_minute', '1500'), -- 15 рублей за минуту
    ('vacuum', 'price_per_minute_with_chemistry', '1500'); -- 15 рублей за минуту (без химии) 