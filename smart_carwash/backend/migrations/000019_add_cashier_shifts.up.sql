-- Создаем таблицу для смен кассиров
CREATE TABLE IF NOT EXISTS cashier_shifts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cashier_id UUID NOT NULL REFERENCES cashiers(id),
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Создаем индексы для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_cashier_shifts_cashier_id ON cashier_shifts(cashier_id);
CREATE INDEX IF NOT EXISTS idx_cashier_shifts_is_active ON cashier_shifts(is_active);
CREATE INDEX IF NOT EXISTS idx_cashier_shifts_expires_at ON cashier_shifts(expires_at);

-- Создаем триггер для автоматического обновления updated_at
CREATE TRIGGER update_cashier_shifts_updated_at
BEFORE UPDATE ON cashier_shifts
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column(); 