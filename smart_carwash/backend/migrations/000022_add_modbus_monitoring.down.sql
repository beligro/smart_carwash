-- Удаление триггера и функции
DROP TRIGGER IF EXISTS trigger_update_modbus_connection_status_updated_at ON modbus_connection_statuses;
DROP FUNCTION IF EXISTS update_modbus_connection_status_updated_at();

-- Удаление индексов
DROP INDEX IF EXISTS idx_modbus_connection_statuses_updated_at;
DROP INDEX IF EXISTS idx_modbus_connection_statuses_connected;
DROP INDEX IF EXISTS idx_modbus_connection_statuses_box_id;

DROP INDEX IF EXISTS idx_modbus_operations_box_created;
DROP INDEX IF EXISTS idx_modbus_operations_operation;
DROP INDEX IF EXISTS idx_modbus_operations_success;
DROP INDEX IF EXISTS idx_modbus_operations_created_at;
DROP INDEX IF EXISTS idx_modbus_operations_box_id;

-- Удаление таблиц
DROP TABLE IF EXISTS modbus_connection_statuses;
DROP TABLE IF EXISTS modbus_operations;


