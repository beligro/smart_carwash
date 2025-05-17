-- Удаление индексов
DROP INDEX IF EXISTS idx_users_telegram_id;
DROP INDEX IF EXISTS idx_wash_boxes_status;

-- Удаление таблиц
DROP TABLE IF EXISTS wash_boxes;
DROP TABLE IF EXISTS users;

-- Удаление расширения uuid-ossp
DROP EXTENSION IF EXISTS "uuid-ossp";
