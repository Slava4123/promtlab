DROP INDEX IF EXISTS idx_subscriptions_paused_resume;
DROP INDEX IF EXISTS idx_subscriptions_user_active;
CREATE UNIQUE INDEX idx_subscriptions_user_active
    ON subscriptions (user_id) WHERE status IN ('active','past_due');

ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS paused_until,
    DROP COLUMN IF EXISTS paused_at;
