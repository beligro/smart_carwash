DROP INDEX IF EXISTS idx_sessions_is_priority;
ALTER TABLE sessions DROP COLUMN IF EXISTS is_priority;
