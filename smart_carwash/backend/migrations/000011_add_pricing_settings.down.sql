-- Удаление настроек цен
DELETE FROM service_settings 
WHERE setting_key IN ('price_per_minute', 'price_per_minute_with_chemistry')
AND service_type IN ('wash', 'air_dry', 'vacuum'); 