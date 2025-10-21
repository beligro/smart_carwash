-- Добавление полей для отслеживания статусов света и химии в боксах
ALTER TABLE modbus_connection_statuses ADD COLUMN light_status BOOLEAN;
ALTER TABLE modbus_connection_statuses ADD COLUMN chemistry_status BOOLEAN;

-- Создание индексов для оптимизации запросов по статусам
CREATE INDEX IF NOT EXISTS idx_modbus_connection_statuses_light_status ON modbus_connection_statuses(light_status) WHERE light_status IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_modbus_connection_statuses_chemistry_status ON modbus_connection_statuses(chemistry_status) WHERE chemistry_status IS NOT NULL;

