-- Удаляем поля cooldown системы из таблицы wash_boxes
ALTER TABLE wash_boxes DROP COLUMN IF EXISTS last_completed_session_user_id;
ALTER TABLE wash_boxes DROP COLUMN IF EXISTS last_completed_at;
ALTER TABLE wash_boxes DROP COLUMN IF EXISTS cooldown_until;

-- Удаляем настройку cooldown_timeout_minutes
DELETE FROM service_settings 
WHERE service_type = 'session' AND setting_key = 'cooldown_timeout_minutes';
