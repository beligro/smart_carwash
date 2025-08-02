-- Удаление индекса
DROP INDEX IF EXISTS idx_payments_session_type;

-- Удаление поля payment_type из таблицы payments
ALTER TABLE payments DROP COLUMN IF EXISTS payment_type; 