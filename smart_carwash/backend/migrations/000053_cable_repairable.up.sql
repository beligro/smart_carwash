-- Кабель пылесоса часто именно ремонтируют (не только меняют) — добавляем опцию «ремонт».
UPDATE component_types SET repairable = true WHERE name = 'Кабель';
