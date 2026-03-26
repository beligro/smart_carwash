-- Добавление полей для программы лояльности "Каждая 10-я мойка бесплатно"

-- Добавляем счетчик завершенных моек с химией
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS completed_washes_count INTEGER DEFAULT 0;

-- Добавляем флаг подписки на канал
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS is_subscribed_to_channel BOOLEAN DEFAULT FALSE;

-- Создаем индекс для быстрого поиска пользователей близких к бесплатной мойке
CREATE INDEX IF NOT EXISTS idx_users_loyalty_count ON users(completed_washes_count) 
WHERE completed_washes_count % 10 >= 8;

-- Добавляем настройки для бесплатной мойки
INSERT INTO service_settings (key, value, description) 
VALUES 
  ('loyalty_free_wash_minutes', '30', 'Время бесплатной мойки (минуты)'),
  ('loyalty_free_chemistry_minutes', '5', 'Время химии для бесплатной мойки (минуты)')
ON CONFLICT (key) DO NOTHING;

-- Добавляем поле для защиты от двойного инкремента
ALTER TABLE sessions
ADD COLUMN IF NOT EXISTS loyalty_counted BOOLEAN DEFAULT FALSE;

-- Комментарии к полям
COMMENT ON COLUMN users.completed_washes_count IS 'Счетчик завершенных моек с химией для программы лояльности';
COMMENT ON COLUMN users.is_subscribed_to_channel IS 'Подписан ли пользователь на канал @h2o_nsk_carwash';
COMMENT ON COLUMN sessions.loyalty_counted IS 'Засчитана ли сессия в программе лояльности (защита от дубликатов)';
