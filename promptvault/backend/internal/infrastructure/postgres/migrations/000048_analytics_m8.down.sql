DROP INDEX IF EXISTS idx_prompts_content_trgm;
-- Extension не дропаем: могут быть другие индексы, зависящие от pg_trgm.
-- CREATE EXTENSION — idempotent; down-up без ошибок.
