DROP INDEX IF EXISTS idx_subscriptions_renewal;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS auto_renew;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS rebill_id;
