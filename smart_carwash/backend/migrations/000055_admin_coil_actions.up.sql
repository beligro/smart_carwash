CREATE TABLE IF NOT EXISTS admin_coil_actions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_username VARCHAR(255) NOT NULL DEFAULT '',
    box_id        UUID,
    box_number    INTEGER NOT NULL,
    box_type      VARCHAR(20) NOT NULL,
    reason        VARCHAR(20) NOT NULL,        -- personal_wash | test
    ticket_id     UUID,
    started_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at    TIMESTAMPTZ,
    ended_at      TIMESTAMPTZ,
    ended_reason  VARCHAR(20),                 -- auto | manual
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_admin_coil_actions_open ON admin_coil_actions (box_id) WHERE ended_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_admin_coil_actions_started ON admin_coil_actions (started_at);
