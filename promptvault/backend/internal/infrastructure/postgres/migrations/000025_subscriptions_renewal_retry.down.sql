DROP INDEX IF EXISTS idx_subscriptions_past_due_retry;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS renewal_attempts;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS last_renewal_attempt_at;
