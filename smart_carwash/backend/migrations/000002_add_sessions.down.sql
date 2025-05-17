-- Удаление индексов
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_box_id;
DROP INDEX IF EXISTS idx_sessions_status;

-- Удаление таблицы сессий
DROP TABLE IF EXISTS sessions;
