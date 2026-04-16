DROP INDEX IF EXISTS idx_user_streaks_at_risk;
ALTER TABLE user_streaks DROP COLUMN IF EXISTS reminder_sent_on;
