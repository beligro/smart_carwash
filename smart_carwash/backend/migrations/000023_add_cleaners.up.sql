-- Добавление таблиц для уборщиков
CREATE TABLE IF NOT EXISTS cleaners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Добавление таблицы сессий уборщиков
CREATE TABLE IF NOT EXISTS cleaner_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cleaner_id UUID NOT NULL,
    token TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (cleaner_id) REFERENCES cleaners(id) ON DELETE CASCADE
);

-- Добавление полей для уборки в таблицу wash_boxes
ALTER TABLE wash_boxes ADD COLUMN IF NOT EXISTS cleaning_reserved_by UUID;
ALTER TABLE wash_boxes ADD COLUMN IF NOT EXISTS cleaning_started_at TIMESTAMP WITH TIME ZONE;

-- Добавление внешнего ключа для cleaning_reserved_by
ALTER TABLE wash_boxes ADD CONSTRAINT fk_washboxes_cleaning_reserved_by 
    FOREIGN KEY (cleaning_reserved_by) REFERENCES cleaners(id) ON DELETE SET NULL;

-- Добавление индексов
CREATE INDEX IF NOT EXISTS idx_cleaners_username ON cleaners(username);
CREATE INDEX IF NOT EXISTS idx_cleaners_is_active ON cleaners(is_active);
CREATE INDEX IF NOT EXISTS idx_cleaner_sessions_cleaner_id ON cleaner_sessions(cleaner_id);
CREATE INDEX IF NOT EXISTS idx_cleaner_sessions_token ON cleaner_sessions(token);
CREATE INDEX IF NOT EXISTS idx_washboxes_cleaning_reserved_by ON wash_boxes(cleaning_reserved_by);
CREATE INDEX IF NOT EXISTS idx_washboxes_cleaning_started_at ON wash_boxes(cleaning_started_at);

-- Добавление настройки времени уборки
INSERT INTO service_settings (service_type, setting_key, setting_value)
VALUES ('cleaner', 'cleaning_timeout_minutes', '30')
ON CONFLICT (service_type, setting_key) DO NOTHING;
