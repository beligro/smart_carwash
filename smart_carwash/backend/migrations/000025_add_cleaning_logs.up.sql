-- Создание таблицы для логов уборки
CREATE TABLE IF NOT EXISTS cleaning_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cleaner_id UUID NOT NULL REFERENCES cleaners(id) ON DELETE CASCADE,
    wash_box_id UUID NOT NULL REFERENCES wash_boxes(id) ON DELETE CASCADE,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    duration_minutes INT,
    status VARCHAR(20) NOT NULL DEFAULT 'in_progress',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Добавление индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_cleaning_logs_cleaner_id ON cleaning_logs(cleaner_id);
CREATE INDEX IF NOT EXISTS idx_cleaning_logs_wash_box_id ON cleaning_logs(wash_box_id);
CREATE INDEX IF NOT EXISTS idx_cleaning_logs_started_at ON cleaning_logs(started_at);
CREATE INDEX IF NOT EXISTS idx_cleaning_logs_status ON cleaning_logs(status);
CREATE INDEX IF NOT EXISTS idx_cleaning_logs_cleaner_started ON cleaning_logs(cleaner_id, started_at);
CREATE INDEX IF NOT EXISTS idx_cleaning_logs_box_started ON cleaning_logs(wash_box_id, started_at);

-- Обновление настройки времени уборки с 30 на 3 минуты
UPDATE service_settings 
SET setting_value = '3' 
WHERE service_type = 'cleaner' AND setting_key = 'cleaning_timeout_minutes';

-- Если настройка не существует, создаем её
INSERT INTO service_settings (service_type, setting_key, setting_value)
VALUES ('cleaner', 'cleaning_timeout_minutes', '3')
ON CONFLICT (service_type, setting_key) DO NOTHING;

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_cleaning_logs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_cleaning_logs_updated_at
    BEFORE UPDATE ON cleaning_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_cleaning_logs_updated_at();

