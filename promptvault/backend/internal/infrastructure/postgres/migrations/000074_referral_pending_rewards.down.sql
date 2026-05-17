-- 000074_referral_pending_rewards.down.sql
DROP INDEX IF EXISTS idx_referral_pending_eligible_at;
DROP INDEX IF EXISTS idx_referral_pending_unique_referee;
DROP TABLE IF EXISTS referral_pending_rewards;
