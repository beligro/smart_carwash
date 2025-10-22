-- Добавить поле для времени химии при продлении сессии
ALTER TABLE sessions ADD COLUMN extension_chemistry_time_minutes INT DEFAULT 0;

-- Добавить комментарий к полю
COMMENT ON COLUMN sessions.extension_chemistry_time_minutes IS 'Время химии, добавленное при продлении сессии (в минутах)';
