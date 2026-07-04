-- Не все детали можно ремонтировать. Ремонт применим только к:
--   шлангам высокого давления (в боксе и в аппаратной), магистральной трубе
--   и электромотору (перемотка). Остальное — только замена.

ALTER TABLE component_types ADD COLUMN IF NOT EXISTS repairable BOOLEAN NOT NULL DEFAULT false;

UPDATE component_types
SET repairable = true
WHERE name IN ('Шланг АВД', 'Шланг в аппаратной', 'Магистральная труба', 'Электромотор');
