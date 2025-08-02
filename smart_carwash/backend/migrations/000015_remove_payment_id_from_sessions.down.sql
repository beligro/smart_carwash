-- Восстановление поля payment_id в таблице sessions
ALTER TABLE sessions ADD COLUMN payment_id UUID;
CREATE INDEX idx_sessions_payment_id ON sessions(payment_id); 