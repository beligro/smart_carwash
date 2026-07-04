DELETE FROM symptom_types WHERE name = 'Бокс заблокирован (не поломка)';
ALTER TABLE service_tickets DROP COLUMN IF EXISTS is_breakdown;
ALTER TABLE symptom_types   DROP COLUMN IF EXISTS is_breakdown;
