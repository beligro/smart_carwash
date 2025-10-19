-- Добавляем поле priority в таблицу wash_boxes
ALTER TABLE wash_boxes ADD COLUMN priority INTEGER NOT NULL DEFAULT 1;

-- Добавляем индекс для оптимизации сортировки по приоритету
CREATE INDEX idx_wash_boxes_priority ON wash_boxes(priority);

-- Добавляем ограничение на минимальное значение приоритета
ALTER TABLE wash_boxes ADD CONSTRAINT chk_wash_boxes_priority_positive CHECK (priority >= 1);
