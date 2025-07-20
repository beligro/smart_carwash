-- Удаление индексов
DROP INDEX IF EXISTS idx_sessions_payment_id;
DROP INDEX IF EXISTS idx_payments_tinkoff_id;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_session_id;

-- Удаление поля payment_id из таблицы sessions
ALTER TABLE sessions DROP COLUMN IF EXISTS payment_id;

-- Удаление таблицы платежей
DROP TABLE IF EXISTS payments; 