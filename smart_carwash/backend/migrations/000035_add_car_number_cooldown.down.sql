-- Удаляем индекс для поиска боксов в кулдауне по госномеру
DROP INDEX IF EXISTS idx_wash_boxes_car_number_cooldown;

-- Удаляем поле для хранения госномера последней завершенной сессии в боксе
ALTER TABLE wash_boxes DROP COLUMN IF EXISTS last_completed_session_car_number;
