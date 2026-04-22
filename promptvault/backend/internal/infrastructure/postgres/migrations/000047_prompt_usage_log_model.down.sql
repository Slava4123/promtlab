DROP INDEX IF EXISTS idx_pul_model_used;
ALTER TABLE prompt_usage_log DROP COLUMN IF EXISTS model_used;
