DROP INDEX IF EXISTS idx_sessions_guest_token;
ALTER TABLE sessions DROP COLUMN IF EXISTS guest_token;
