DROP INDEX IF EXISTS idx_wash_boxes_service_until;
ALTER TABLE wash_boxes DROP COLUMN IF EXISTS service_until;
