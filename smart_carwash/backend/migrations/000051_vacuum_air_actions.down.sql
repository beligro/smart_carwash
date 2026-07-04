UPDATE component_types
SET group_id = (SELECT id FROM component_groups WHERE name = 'Общее' ORDER BY id LIMIT 1)
WHERE name = 'Замена масла';
DELETE FROM component_groups WHERE name = 'ТО аппарата';
ALTER TABLE component_types DROP COLUMN IF EXISTS cleanable;
