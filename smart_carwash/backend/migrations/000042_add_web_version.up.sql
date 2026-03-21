-- Веб-версия: пользователи по телефону, источник сессии, токены привязки Telegram

-- users: телефон и верификация (для входа с веба)
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20) UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMP WITH TIME ZONE;
-- telegram_id делаем nullable для веб-пользователей без привязки
ALTER TABLE users ALTER COLUMN telegram_id DROP NOT NULL;

-- sessions: источник создания (web | telegram)
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS source VARCHAR(10) NOT NULL DEFAULT 'telegram';

-- Таблица одноразовых токенов привязки Telegram (link_tokens)
CREATE TABLE IF NOT EXISTS link_tokens (
    token VARCHAR(32) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_link_tokens_user_id ON link_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_link_tokens_expires_at ON link_tokens(expires_at);

