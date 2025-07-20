-- Добавляем настройки цен для всех типов услуг
INSERT INTO service_settings (service_type, setting_key, setting_value) VALUES 
    ('wash', 'price_per_minute', '3000'),           -- 30 руб/мин
    ('wash', 'chemistry_price_per_minute', '2000'), -- 20 руб/мин за химию
    ('air_dry', 'price_per_minute', '2000'),        -- 20 руб/мин
    ('air_dry', 'chemistry_price_per_minute', '1000'), -- 10 руб/мин за химию
    ('vacuum', 'price_per_minute', '1500'),         -- 15 руб/мин
    ('vacuum', 'chemistry_price_per_minute', '500');  -- 5 руб/мин за химию 