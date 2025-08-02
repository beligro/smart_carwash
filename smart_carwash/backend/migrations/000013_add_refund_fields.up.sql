-- Добавление полей для возврата денег в таблицу payments
ALTER TABLE payments ADD COLUMN refunded_amount INTEGER DEFAULT 0;
ALTER TABLE payments ADD COLUMN refunded_at TIMESTAMP WITH TIME ZONE;

-- Добавление индекса для поля refunded_amount для быстрого поиска возвращенных платежей
CREATE INDEX IF NOT EXISTS idx_payments_refunded_amount ON payments(refunded_amount); 