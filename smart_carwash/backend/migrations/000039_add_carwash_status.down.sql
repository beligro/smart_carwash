-- Откат миграции для статуса мойки

-- Удаление индексов
DROP INDEX IF EXISTS idx_carwash_status_history_created_by;
DROP INDEX IF EXISTS idx_carwash_status_history_created_at;
DROP INDEX IF EXISTS idx_carwash_status_is_closed;

-- Удаление таблиц
DROP TABLE IF EXISTS carwash_status_history;
DROP TABLE IF EXISTS carwash_status;




