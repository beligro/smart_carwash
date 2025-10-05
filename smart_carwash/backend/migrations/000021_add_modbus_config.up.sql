-- Добавление полей для Modbus конфигурации (регистры для каждого бокса)
ALTER TABLE wash_boxes ADD COLUMN light_coil_register VARCHAR(10);
ALTER TABLE wash_boxes ADD COLUMN chemistry_coil_register VARCHAR(10);

-- Добавление индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_wash_boxes_light_coil ON wash_boxes(light_coil_register);
CREATE INDEX IF NOT EXISTS idx_wash_boxes_chemistry_coil ON wash_boxes(chemistry_coil_register); 