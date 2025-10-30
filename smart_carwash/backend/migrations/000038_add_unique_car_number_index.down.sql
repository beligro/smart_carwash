-- Удаление уникального индекса для активных сессий с номером машины
DROP INDEX IF EXISTS idx_sessions_car_number_unique_active;
