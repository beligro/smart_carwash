ALTER TABLE sessions DROP COLUMN IF EXISTS active_started_at;
ALTER TABLE sessions DROP COLUMN IF EXISTS active_ended_at;
ALTER TABLE sessions DROP COLUMN IF EXISTS completion_source;
ALTER TABLE sessions DROP COLUMN IF EXISTS completion_user_id;
