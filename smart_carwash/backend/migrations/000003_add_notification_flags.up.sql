-- Добавляем флаги для отслеживания отправленных уведомлений
ALTER TABLE sessions 
ADD COLUMN is_expiring_notification_sent BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN is_completing_notification_sent BOOLEAN NOT NULL DEFAULT FALSE;
