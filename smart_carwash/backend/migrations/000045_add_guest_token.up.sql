ALTER TABLE sessions ADD COLUMN IF NOT EXISTS guest_token VARCHAR(64) UNIQUE NULL;
CREATE INDEX IF NOT EXISTS idx_sessions_guest_token ON sessions(guest_token) WHERE guest_token IS NOT NULL;

-- Служебный гостевой пользователь: все гостевые сессии ссылаются на него (FK user_id),
-- а конкретные гости различаются по guest_token. Аналог кассирского пользователя.
INSERT INTO users (id, first_name, created_at, updated_at)
VALUES ('11111111-1111-1111-1111-111111111111', 'Guest', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
