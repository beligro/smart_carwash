-- Удаляем триггеры
DROP TRIGGER IF EXISTS update_cashiers_updated_at ON cashiers;
DROP TRIGGER IF EXISTS update_cashier_sessions_updated_at ON cashier_sessions;
DROP TRIGGER IF EXISTS update_two_factor_auth_settings_updated_at ON two_factor_auth_settings;

-- Удаляем индексы
DROP INDEX IF EXISTS idx_cashier_sessions_token;

-- Удаляем таблицы
DROP TABLE IF EXISTS cashier_sessions;
DROP TABLE IF EXISTS cashiers;
DROP TABLE IF EXISTS two_factor_auth_settings;

-- Удаляем функцию триггера, если она больше не используется
-- (проверяем, есть ли еще триггеры, использующие эту функцию)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger
        WHERE tgname IN (
            'update_users_updated_at',
            'update_washboxes_updated_at',
            'update_sessions_updated_at'
        )
    ) THEN
        DROP FUNCTION IF EXISTS update_updated_at_column();
    END IF;
END
$$;
