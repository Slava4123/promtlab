ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE users DROP COLUMN IF EXISTS reengagement_sent_at;
ALTER TABLE users DROP COLUMN IF EXISTS welcome_sent_at;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS pre_expire_stage;
