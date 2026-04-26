DROP INDEX IF EXISTS idx_prompts_slug_aliases_gin;

ALTER TABLE prompts DROP COLUMN IF EXISTS slug_aliases;
