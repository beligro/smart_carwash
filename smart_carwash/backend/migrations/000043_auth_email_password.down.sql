-- Откат: возврат phone, удаление password_hash и email_verified_at

ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20) UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;

DROP INDEX IF EXISTS idx_users_email_unique;
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
