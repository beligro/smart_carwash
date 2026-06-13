-- Авторизация по почте: password_hash, email_verified_at; удаление phone

-- users: добавляем поля для входа по email
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP WITH TIME ZONE;

-- Уникальный индекс по email для непустых значений (веб-пользователи)
DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users(email) WHERE email IS NOT NULL AND email != '';

-- Удаление полей телефона
ALTER TABLE users DROP COLUMN IF EXISTS phone_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS phone;
