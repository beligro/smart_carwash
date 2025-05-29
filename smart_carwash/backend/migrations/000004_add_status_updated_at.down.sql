-- Удаляем поле для отслеживания времени последнего обновления статуса
ALTER TABLE sessions 
DROP COLUMN status_updated_at;
