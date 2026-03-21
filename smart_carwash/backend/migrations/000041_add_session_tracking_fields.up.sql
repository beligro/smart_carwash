-- Поля для отслеживания жизненного цикла сессии:
-- кто и когда начал мойку, кто и как завершил

ALTER TABLE sessions ADD COLUMN IF NOT EXISTS active_started_at TIMESTAMPTZ NULL;
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS active_ended_at TIMESTAMPTZ NULL;
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS completion_source VARCHAR(20) NULL;
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS completion_user_id UUID NULL;
