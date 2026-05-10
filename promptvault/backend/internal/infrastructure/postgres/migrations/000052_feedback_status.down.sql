DROP INDEX IF EXISTS idx_feedbacks_status;
ALTER TABLE feedbacks DROP COLUMN IF EXISTS status;
DROP TYPE IF EXISTS feedback_status;
