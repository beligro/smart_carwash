-- Добавление времени мойки в таблицу sessions
ALTER TABLE sessions ADD COLUMN rental_time_minutes INT DEFAULT 5;

-- Создание таблицы настроек для времени мойки
CREATE TABLE IF NOT EXISTS service_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_type VARCHAR(50) NOT NULL,
    setting_key VARCHAR(50) NOT NULL,
    setting_value JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(service_type, setting_key)
);

-- Добавление индексов
CREATE INDEX IF NOT EXISTS idx_service_settings_service_type ON service_settings(service_type);
CREATE INDEX IF NOT EXISTS idx_service_settings_setting_key ON service_settings(setting_key);

-- Добавление начальных данных для времени мойки
INSERT INTO service_settings (service_type, setting_key, setting_value)
VALUES 
    ('wash', 'available_rental_times', '[5, 10, 15, 20]'),
    ('air_dry', 'available_rental_times', '[3, 5, 10]'),
    ('vacuum', 'available_rental_times', '[5, 10, 15]');
