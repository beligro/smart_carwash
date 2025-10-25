-- Удаляем поле email из таблицы sessions
ALTER TABLE sessions DROP COLUMN email;

-- Удаляем поле email из таблицы users
ALTER TABLE users DROP COLUMN email;

-- Удаляем индекс для email в users
DROP INDEX IF EXISTS idx_users_email;
