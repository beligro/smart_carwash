-- Добавление типа платежа в таблицу payments
ALTER TABLE payments ADD COLUMN payment_type VARCHAR(20) DEFAULT 'main';

-- Обновляем существующие платежи
UPDATE payments SET payment_type = 'main' WHERE payment_type IS NULL;

-- Создаем индекс для производительности
CREATE INDEX idx_payments_session_type ON payments(session_id, payment_type); 