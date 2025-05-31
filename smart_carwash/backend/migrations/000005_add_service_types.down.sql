-- Удаление типа услуги и флага химии из таблицы sessions
ALTER TABLE sessions DROP COLUMN service_type;
ALTER TABLE sessions DROP COLUMN with_chemistry;

-- Удаление типа бокса из таблицы wash_boxes
ALTER TABLE wash_boxes DROP COLUMN service_type;
