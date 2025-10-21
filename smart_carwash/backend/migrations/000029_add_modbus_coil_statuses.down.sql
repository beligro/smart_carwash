-- Откат миграции - удаление индексов и полей
DROP INDEX IF EXISTS idx_modbus_connection_statuses_chemistry_status;
DROP INDEX IF EXISTS idx_modbus_connection_statuses_light_status;

ALTER TABLE modbus_connection_statuses DROP COLUMN IF EXISTS chemistry_status;
ALTER TABLE modbus_connection_statuses DROP COLUMN IF EXISTS light_status;

