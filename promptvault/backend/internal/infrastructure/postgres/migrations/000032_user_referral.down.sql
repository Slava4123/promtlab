DROP INDEX IF EXISTS idx_users_referred_by;
DROP INDEX IF EXISTS idx_users_referral_code;
ALTER TABLE users
    DROP COLUMN IF EXISTS referral_rewarded_at,
    DROP COLUMN IF EXISTS referred_by,
    DROP COLUMN IF EXISTS referral_code;
