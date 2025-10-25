-- Добавляем поле email в таблицу users
ALTER TABLE users ADD COLUMN email VARCHAR(255);

-- Добавляем поле email в таблицу sessions
ALTER TABLE sessions ADD COLUMN email VARCHAR(255);

-- Создаем индекс для email в users (опционально, для быстрого поиска)
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
