ALTER TABLE sessions ADD COLUMN is_priority BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX idx_sessions_is_priority ON sessions(is_priority);
