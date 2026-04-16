DROP INDEX IF EXISTS idx_prompts_content_trgm;
DROP INDEX IF EXISTS idx_prompts_title_trgm;
DROP INDEX IF EXISTS idx_prompts_team_updated;
DROP INDEX IF EXISTS idx_prompts_user_updated;
-- pg_trgm extension не удаляем в down — может использоваться другими индексами
-- или будущими миграциями.
