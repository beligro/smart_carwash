-- Удаление индексов
DROP INDEX IF EXISTS idx_wash_boxes_light_coil;
DROP INDEX IF EXISTS idx_wash_boxes_chemistry_coil;

-- Удаление полей Modbus конфигурации
ALTER TABLE wash_boxes DROP COLUMN IF EXISTS light_coil_register;
ALTER TABLE wash_boxes DROP COLUMN IF EXISTS chemistry_coil_register; 