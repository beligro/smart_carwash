-- Добавляем поле для отслеживания времени последнего обновления статуса
ALTER TABLE sessions 
ADD COLUMN status_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();

-- Инициализируем поле текущим значением updated_at
UPDATE sessions SET status_updated_at = updated_at;
