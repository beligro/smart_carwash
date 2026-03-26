-- Откат миграции программы лояльности

-- Удаляем настройки
DELETE FROM service_settings WHERE key IN ('loyalty_free_wash_minutes', 'loyalty_free_chemistry_minutes');

-- Удаляем индекс
DROP INDEX IF EXISTS idx_users_loyalty_count;

-- Удаляем поля
ALTER TABLE sessions DROP COLUMN IF EXISTS loyalty_counted;
ALTER TABLE users DROP COLUMN IF EXISTS is_subscribed_to_channel;
ALTER TABLE users DROP COLUMN IF EXISTS completed_washes_count;
