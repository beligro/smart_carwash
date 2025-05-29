-- Удаляем флаги для отслеживания отправленных уведомлений
ALTER TABLE sessions 
DROP COLUMN is_expiring_notification_sent,
DROP COLUMN is_completing_notification_sent;
