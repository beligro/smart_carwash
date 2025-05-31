-- Добавление типа бокса в таблицу wash_boxes
ALTER TABLE wash_boxes ADD COLUMN service_type VARCHAR(50) NOT NULL DEFAULT 'wash';

-- Добавление типа услуги и флага химии в таблицу sessions
ALTER TABLE sessions ADD COLUMN service_type VARCHAR(50);
ALTER TABLE sessions ADD COLUMN with_chemistry BOOLEAN DEFAULT FALSE;

-- Обновление существующих боксов
-- Первая треть боксов - мойка
UPDATE wash_boxes SET service_type = 'wash' WHERE id IN (
    SELECT id FROM wash_boxes ORDER BY number LIMIT (SELECT COUNT(*)/3 FROM wash_boxes)
);

-- Вторая треть боксов - обдув
UPDATE wash_boxes SET service_type = 'air_dry' WHERE id IN (
    SELECT id FROM wash_boxes WHERE service_type = 'wash' ORDER BY number LIMIT (SELECT COUNT(*)/3 FROM wash_boxes) OFFSET (SELECT COUNT(*)/3 FROM wash_boxes)
);

-- Последняя треть боксов - пылесос
UPDATE wash_boxes SET service_type = 'vacuum' WHERE service_type = 'wash';
