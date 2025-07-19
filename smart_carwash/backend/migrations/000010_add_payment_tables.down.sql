-- Удаляем триггеры
DROP TRIGGER IF EXISTS update_payment_events_updated_at ON payment_events;
DROP TRIGGER IF EXISTS update_payment_refunds_updated_at ON payment_refunds;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;

-- Удаляем индексы
DROP INDEX IF EXISTS idx_payment_events_created_at;
DROP INDEX IF EXISTS idx_payment_events_event_type;
DROP INDEX IF EXISTS idx_payment_events_payment_id;

DROP INDEX IF EXISTS idx_payment_refunds_next_retry;
DROP INDEX IF EXISTS idx_payment_refunds_status;
DROP INDEX IF EXISTS idx_payment_refunds_payment_id;

DROP INDEX IF EXISTS idx_payments_tinkoff_id;
DROP INDEX IF EXISTS idx_payments_created_at;
DROP INDEX IF EXISTS idx_payments_type;
DROP INDEX IF EXISTS idx_payments_status;
DROP INDEX IF EXISTS idx_payments_session_id;
DROP INDEX IF EXISTS idx_payments_user_id;

-- Удаляем таблицы
DROP TABLE IF EXISTS payment_events;
DROP TABLE IF EXISTS payment_refunds;
DROP TABLE IF EXISTS payments; 