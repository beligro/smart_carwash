-- Добавление поля chemistry_enabled в таблицу wash_boxes
ALTER TABLE wash_boxes ADD COLUMN chemistry_enabled BOOLEAN DEFAULT TRUE;

-- Устанавливаем значения по умолчанию
UPDATE wash_boxes SET chemistry_enabled = TRUE WHERE service_type = 'wash';
UPDATE wash_boxes SET chemistry_enabled = FALSE WHERE service_type IN ('air_dry', 'vacuum'); 