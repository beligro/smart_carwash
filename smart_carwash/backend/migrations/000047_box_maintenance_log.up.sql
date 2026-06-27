-- Журнал отметок ТО боксов-моек.
-- Каждая запись фиксирует факт сброса счётчика моточасов (плановое ТО, ремонт и т.п.).

CREATE TABLE IF NOT EXISTS box_maintenance_log (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    box_number           INTEGER NOT NULL,
    performed_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    performed_by         VARCHAR(255),
    motor_hours_at_reset INTEGER,
    reason               VARCHAR(100),
    comment              TEXT
);

CREATE INDEX IF NOT EXISTS idx_box_maintenance_log_box_performed_at
    ON box_maintenance_log (box_number, performed_at DESC);
