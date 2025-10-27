-- Добавляем поле для хранения госномера последней завершенной сессии в боксе
ALTER TABLE wash_boxes ADD COLUMN last_completed_session_car_number VARCHAR(20);

-- Добавляем индекс для быстрого поиска боксов в кулдауне по госномеру
CREATE INDEX idx_wash_boxes_car_number_cooldown ON wash_boxes(last_completed_session_car_number, cooldown_until) 
WHERE last_completed_session_car_number IS NOT NULL AND cooldown_until IS NOT NULL;
