-- Таймерный сервис бокса (для чистки пылесосов): бокс переводится в maintenance
-- на заданное время и автоматически возвращается в работу по истечении service_until.
-- Не наряд: без симптома, без push, без статистики поломок.

ALTER TABLE wash_boxes ADD COLUMN IF NOT EXISTS service_until TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_wash_boxes_service_until ON wash_boxes (service_until)
    WHERE service_until IS NOT NULL;
