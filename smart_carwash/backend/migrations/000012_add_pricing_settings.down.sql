-- Удаляем настройки цен
DELETE FROM service_settings WHERE setting_key IN ('price_per_minute', 'chemistry_price_per_minute'); 