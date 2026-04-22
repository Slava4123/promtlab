DROP INDEX IF EXISTS idx_prompt_versions_changed_by;
ALTER TABLE prompt_versions DROP COLUMN IF EXISTS changed_by;
