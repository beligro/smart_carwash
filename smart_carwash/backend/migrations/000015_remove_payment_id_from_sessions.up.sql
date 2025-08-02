-- Удаление поля payment_id из таблицы sessions
ALTER TABLE sessions DROP COLUMN IF EXISTS payment_id; 