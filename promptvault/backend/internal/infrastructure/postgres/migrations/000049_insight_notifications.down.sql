ALTER TABLE users DROP COLUMN IF EXISTS insight_emails_enabled;
DROP INDEX IF EXISTS idx_insight_notif_user_type_time;
DROP TABLE IF EXISTS insight_notifications;
