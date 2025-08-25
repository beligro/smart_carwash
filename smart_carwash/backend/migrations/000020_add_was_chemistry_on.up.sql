-- Добавление поля was_chemistry_on в таблицу sessions
ALTER TABLE sessions ADD COLUMN was_chemistry_on BOOLEAN DEFAULT FALSE; 