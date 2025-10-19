-- Удаляем ограничение на приоритет
ALTER TABLE wash_boxes DROP CONSTRAINT IF EXISTS chk_wash_boxes_priority_positive;

-- Удаляем индекс
DROP INDEX IF EXISTS idx_wash_boxes_priority;

-- Удаляем поле priority
ALTER TABLE wash_boxes DROP COLUMN IF EXISTS priority;
