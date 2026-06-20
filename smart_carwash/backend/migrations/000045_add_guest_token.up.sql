ALTER TABLE sessions ADD COLUMN IF NOT EXISTS guest_token VARCHAR(64) UNIQUE NULL;
CREATE INDEX IF NOT EXISTS idx_sessions_guest_token ON sessions(guest_token) WHERE guest_token IS NOT NULL;
