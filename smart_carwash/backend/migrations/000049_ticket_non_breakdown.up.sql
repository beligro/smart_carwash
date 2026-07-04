-- Не-поломочные причины постановки в сервис.
-- Пример: бокс заблокирован (заглохшая машина и т.п.) — это не поломка оборудования,
-- в статистику поломок не идёт и при закрытии не требует выбора работ.

ALTER TABLE symptom_types   ADD COLUMN IF NOT EXISTS is_breakdown BOOLEAN NOT NULL DEFAULT true;
ALTER TABLE service_tickets ADD COLUMN IF NOT EXISTS is_breakdown BOOLEAN NOT NULL DEFAULT true;

-- Симптом доступен для любого типа бокса (box_type='any'), не поломка.
INSERT INTO symptom_types (group_name, name, box_type, sort_order, is_breakdown, is_active)
VALUES ('Прочее', 'Бокс заблокирован (не поломка)', 'any', 100, false, true);
