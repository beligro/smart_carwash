-- Добавляем настройки цен для всех типов услуг
INSERT INTO service_settings (service_type, setting_key, setting_value) VALUES 
    ('wash', 'chemistry_price_per_minute', '2000'); -- 20 руб/мин за химию