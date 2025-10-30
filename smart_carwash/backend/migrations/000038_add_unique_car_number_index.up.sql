-- Добавление уникального индекса для предотвращения дублирования активных сессий с одинаковым номером машины
-- Индекс действует только для активных сессий (created, in_queue, assigned, active)

-- Создаем частичный уникальный индекс
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_car_number_unique_active 
ON sessions (car_number) 
WHERE car_number IS NOT NULL 
  AND car_number != '' 
  AND status IN ('created', 'in_queue', 'assigned', 'active');
