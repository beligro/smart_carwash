-- Уточнение действий по деталям пылесосов/воздуха:
--  - шланг пылесоса: чистка / ремонт / замена;
--  - шланг воздуха: ремонт / замена;
--  - замены масла у пылесоса и воздуха нет (масло — только у моечного аппарата).

ALTER TABLE component_types ADD COLUMN IF NOT EXISTS cleanable BOOLEAN NOT NULL DEFAULT false;

UPDATE component_types SET repairable = true WHERE name IN ('Шланг пылесоса', 'Шланг воздуха');
UPDATE component_types SET cleanable  = true WHERE name = 'Шланг пылесоса';

-- «Замена масла» переносим из общей группы (any) в группу только для мойки.
INSERT INTO component_groups (name, carrier, box_type, sort_order)
VALUES ('ТО аппарата', 'machine', 'wash', 12);

UPDATE component_types
SET group_id = (SELECT id FROM component_groups WHERE name = 'ТО аппарата' ORDER BY id DESC LIMIT 1)
WHERE name = 'Замена масла';
