-- Удаляем базовые цены для услуг
DELETE FROM service_settings WHERE setting_key IN ('price_per_minute', 'chemistry_price_per_minute', 'available_rental_times'); 