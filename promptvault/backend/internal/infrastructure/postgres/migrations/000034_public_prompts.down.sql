DROP INDEX IF EXISTS idx_prompts_public;
DROP INDEX IF EXISTS idx_prompts_slug;
ALTER TABLE prompts
    DROP COLUMN IF EXISTS slug,
    DROP COLUMN IF EXISTS is_public;
