-- Удаление записей с нулевым UUID (битых записей)
DELETE FROM modbus_operations WHERE id = '00000000-0000-0000-0000-000000000000';

-- Также удалим записи где id IS NULL, если такие есть
DELETE FROM modbus_operations WHERE id IS NULL;

