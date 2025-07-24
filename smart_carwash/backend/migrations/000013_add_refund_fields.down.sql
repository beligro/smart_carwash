-- Удаление индекса для поля refunded_amount
DROP INDEX IF EXISTS idx_payments_refunded_amount;

-- Удаление полей для возврата денег из таблицы payments
ALTER TABLE payments DROP COLUMN IF EXISTS refunded_at;
ALTER TABLE payments DROP COLUMN IF EXISTS refunded_amount; 