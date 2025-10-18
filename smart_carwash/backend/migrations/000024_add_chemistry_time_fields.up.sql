-- Добавить поля для времени химии в таблицу sessions
ALTER TABLE sessions ADD COLUMN chemistry_time_minutes INT DEFAULT 0;
ALTER TABLE sessions ADD COLUMN chemistry_started_at TIMESTAMP NULL;
ALTER TABLE sessions ADD COLUMN chemistry_ended_at TIMESTAMP NULL;

-- Добавить индекс для поиска активных химий
CREATE INDEX idx_sessions_chemistry_active ON sessions(chemistry_started_at, chemistry_ended_at);

-- Добавить настройки доступного времени химии по умолчанию (3, 4, 5 минут)
INSERT INTO service_settings (service_type, setting_key, setting_value)
VALUES ('wash', 'available_chemistry_times', '[3, 4, 5]')
ON CONFLICT (service_type, setting_key) DO NOTHING;


