-- Добавляем поля для cooldown системы в таблицу wash_boxes
ALTER TABLE wash_boxes ADD COLUMN last_completed_session_user_id UUID;
ALTER TABLE wash_boxes ADD COLUMN last_completed_at TIMESTAMP;
ALTER TABLE wash_boxes ADD COLUMN cooldown_until TIMESTAMP;

-- Добавляем индексы для оптимизации запросов
CREATE INDEX idx_wash_boxes_last_completed_user ON wash_boxes(last_completed_session_user_id);
CREATE INDEX idx_wash_boxes_cooldown_until ON wash_boxes(cooldown_until);

-- Добавляем настройку cooldown_timeout_minutes по умолчанию
INSERT INTO service_settings (service_type, setting_key, setting_value, created_at, updated_at)
VALUES ('session', 'cooldown_timeout_minutes', '5', NOW(), NOW())
ON CONFLICT (service_type, setting_key) DO NOTHING;
