-- Добавление типа бокса в таблицу wash_boxes
ALTER TABLE wash_boxes ADD COLUMN service_type VARCHAR(50) NOT NULL DEFAULT 'wash';

-- Добавление типа услуги и флага химии в таблицу sessions
ALTER TABLE sessions ADD COLUMN service_type VARCHAR(50);
ALTER TABLE sessions ADD COLUMN with_chemistry BOOLEAN DEFAULT FALSE;
