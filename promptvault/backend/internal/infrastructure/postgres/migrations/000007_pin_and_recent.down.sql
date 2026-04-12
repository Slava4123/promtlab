DROP INDEX IF EXISTS idx_prompts_last_used_at;
ALTER TABLE prompts DROP COLUMN IF EXISTS last_used_at;
DROP TABLE IF EXISTS prompt_pins;
