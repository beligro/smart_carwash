-- Откат изменений: возвращаем числовой тип для priority

-- Удаляем constraint
ALTER TABLE wash_boxes DROP CONSTRAINT IF EXISTS wash_boxes_priority_check;

-- Меняем тип обратно на integer, все значения устанавливаем в 1
ALTER TABLE wash_boxes ALTER COLUMN priority TYPE INTEGER USING 1;

-- Устанавливаем дефолтное значение 1
ALTER TABLE wash_boxes ALTER COLUMN priority SET DEFAULT 1;

-- Восстанавливаем старый constraint
ALTER TABLE wash_boxes ADD CONSTRAINT wash_boxes_priority_check CHECK (priority >= 1);

