-- Создание таблицы для истории изменений боксов
CREATE TABLE IF NOT EXISTS washbox_change_logs (
    id UUID PRIMARY KEY,
    box_id UUID NOT NULL REFERENCES wash_boxes(id) ON DELETE CASCADE,
    box_number INTEGER NOT NULL,
    action VARCHAR(64) NOT NULL,
    prev_status VARCHAR(32),
    new_status VARCHAR(32),
    prev_value BOOLEAN,
    new_value BOOLEAN,
    actor_type VARCHAR(32) NOT NULL,
    source VARCHAR(64),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_washbox_change_logs_box_number ON washbox_change_logs(box_number);
CREATE INDEX IF NOT EXISTS idx_washbox_change_logs_actor_type ON washbox_change_logs(actor_type);
CREATE INDEX IF NOT EXISTS idx_washbox_change_logs_action ON washbox_change_logs(action);
CREATE INDEX IF NOT EXISTS idx_washbox_change_logs_created_at ON washbox_change_logs(created_at);



