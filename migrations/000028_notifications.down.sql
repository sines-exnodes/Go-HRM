-- migrations/000028_notifications.down.sql
--
-- Dropping the table drops its indexes, so they need no explicit statements.
DROP TRIGGER IF EXISTS trg_notifications_set_updated_at ON notifications;
DROP TABLE IF EXISTS notifications;
