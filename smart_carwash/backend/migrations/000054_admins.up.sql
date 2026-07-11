-- Персональные учётки администраторов (Макс/Костя/Стас и др.).
-- Доступ к разделам админки задаётся списком в allowed_sections (через запятую).
CREATE TABLE IF NOT EXISTS admins (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username         VARCHAR(255) NOT NULL,
    password_hash    VARCHAR(255) NOT NULL,
    display_name     VARCHAR(255) NOT NULL DEFAULT '',
    allowed_sections TEXT         NOT NULL DEFAULT '',
    is_active        BOOLEAN      NOT NULL DEFAULT true,
    last_login       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_admins_username_lower ON admins (LOWER(username)) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_admins_deleted_at ON admins (deleted_at);
