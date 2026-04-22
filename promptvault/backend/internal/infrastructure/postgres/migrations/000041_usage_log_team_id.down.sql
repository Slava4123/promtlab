DROP INDEX IF EXISTS idx_usage_log_prompt_time;
CREATE INDEX IF NOT EXISTS idx_usage_log_prompt ON prompt_usage_log (prompt_id);

DROP INDEX IF EXISTS idx_pul_team_used;
ALTER TABLE prompt_usage_log DROP COLUMN IF EXISTS team_id;
