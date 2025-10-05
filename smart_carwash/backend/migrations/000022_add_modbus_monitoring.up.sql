-- Создание таблицы для истории операций Modbus
CREATE TABLE IF NOT EXISTS modbus_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    box_id UUID NOT NULL REFERENCES wash_boxes(id) ON DELETE CASCADE,
    operation VARCHAR(50) NOT NULL,
    register VARCHAR(10) NOT NULL,
    value BOOLEAN NOT NULL,
    success BOOLEAN NOT NULL DEFAULT false,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для статуса подключения Modbus
CREATE TABLE IF NOT EXISTS modbus_connection_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    box_id UUID NOT NULL UNIQUE REFERENCES wash_boxes(id) ON DELETE CASCADE,
    connected BOOLEAN NOT NULL DEFAULT false,
    last_error TEXT,
    last_seen TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_modbus_operations_box_id ON modbus_operations(box_id);
CREATE INDEX IF NOT EXISTS idx_modbus_operations_created_at ON modbus_operations(created_at);
CREATE INDEX IF NOT EXISTS idx_modbus_operations_success ON modbus_operations(success);
CREATE INDEX IF NOT EXISTS idx_modbus_operations_operation ON modbus_operations(operation);
CREATE INDEX IF NOT EXISTS idx_modbus_operations_box_created ON modbus_operations(box_id, created_at);

CREATE INDEX IF NOT EXISTS idx_modbus_connection_statuses_box_id ON modbus_connection_statuses(box_id);
CREATE INDEX IF NOT EXISTS idx_modbus_connection_statuses_connected ON modbus_connection_statuses(connected);
CREATE INDEX IF NOT EXISTS idx_modbus_connection_statuses_updated_at ON modbus_connection_statuses(updated_at);

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_modbus_connection_status_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_modbus_connection_status_updated_at
    BEFORE UPDATE ON modbus_connection_statuses
    FOR EACH ROW
    EXECUTE FUNCTION update_modbus_connection_status_updated_at();


