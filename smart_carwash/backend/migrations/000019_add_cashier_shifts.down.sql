-- Удаляем триггер
DROP TRIGGER IF EXISTS update_cashier_shifts_updated_at ON cashier_shifts;

-- Удаляем индексы
DROP INDEX IF EXISTS idx_cashier_shifts_expires_at;
DROP INDEX IF EXISTS idx_cashier_shifts_is_active;
DROP INDEX IF EXISTS idx_cashier_shifts_cashier_id;

-- Удаляем таблицу
DROP TABLE IF EXISTS cashier_shifts; 