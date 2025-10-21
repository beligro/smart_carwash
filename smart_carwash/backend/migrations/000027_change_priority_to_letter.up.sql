-- Удаляем старый constraint
ALTER TABLE wash_boxes DROP CONSTRAINT IF EXISTS wash_boxes_priority_check;

-- Удаляем старую колонку priority
ALTER TABLE wash_boxes DROP COLUMN priority;

-- Создаем новую колонку priority с типом VARCHAR(1)
ALTER TABLE wash_boxes ADD COLUMN priority VARCHAR(1) NOT NULL DEFAULT 'A';

-- Устанавливаем для всех существующих боксов приоритет 'A'
UPDATE wash_boxes SET priority = 'A';

-- Добавляем constraint для проверки (только заглавные буквы A-Z)
ALTER TABLE wash_boxes ADD CONSTRAINT wash_boxes_priority_check CHECK (priority ~ '^[A-Z]$');

