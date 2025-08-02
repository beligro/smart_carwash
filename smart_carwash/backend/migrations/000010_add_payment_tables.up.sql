-- Создание таблицы платежей
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL,
    amount INTEGER NOT NULL, -- сумма в копейках
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, succeeded, failed
    payment_url TEXT,
    tinkoff_id VARCHAR(255), -- ID платежа в Tinkoff
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

-- Добавление поля payment_id в таблицу sessions
ALTER TABLE sessions ADD COLUMN payment_id UUID REFERENCES payments(id);

-- Добавление индексов
CREATE INDEX IF NOT EXISTS idx_payments_session_id ON payments(session_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_tinkoff_id ON payments(tinkoff_id);
CREATE INDEX IF NOT EXISTS idx_sessions_payment_id ON sessions(payment_id); 