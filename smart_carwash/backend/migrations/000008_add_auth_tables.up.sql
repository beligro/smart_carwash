-- Создаем таблицу для кассиров
CREATE TABLE IF NOT EXISTS cashiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Создаем таблицу для сессий кассиров
CREATE TABLE IF NOT EXISTS cashier_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cashier_id UUID NOT NULL REFERENCES cashiers(id),
    token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Создаем индекс для быстрого поиска сессий по токену
CREATE INDEX IF NOT EXISTS idx_cashier_sessions_token ON cashier_sessions(token);

-- Создаем таблицу для настроек двухфакторной аутентификации
CREATE TABLE IF NOT EXISTS two_factor_auth_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    secret VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- Создаем триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Применяем триггер к таблице кассиров
CREATE TRIGGER update_cashiers_updated_at
BEFORE UPDATE ON cashiers
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Применяем триггер к таблице сессий кассиров
CREATE TRIGGER update_cashier_sessions_updated_at
BEFORE UPDATE ON cashier_sessions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Применяем триггер к таблице настроек двухфакторной аутентификации
CREATE TRIGGER update_two_factor_auth_settings_updated_at
BEFORE UPDATE ON two_factor_auth_settings
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
