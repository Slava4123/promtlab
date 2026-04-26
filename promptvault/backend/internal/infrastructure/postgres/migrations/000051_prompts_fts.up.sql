-- Phase 15 W3: PostgreSQL Full-Text Search для prompts.
--
-- Замена ILIKE %q% на tsvector + GIN. Поддержка русских словоформ
-- (Snowball stemmer), английских словоформ, ё/е симметрия (unaccent).
--
-- Один tsvector с конкатенацией двух конфигов:
--   russian_unaccent stemmer обрабатывает кириллицу + ё/е норм через unaccent;
--   english stemmer обрабатывает латиницу.
-- Snowball-стеммеры пропускают чужие алфавиты — duplicate tokens PG сольёт.
--
-- setweight A для title, B для content — title ранжируется выше при ts_rank_cd.
--
-- GENERATED ALWAYS AS STORED (PG12+) вместо триггера: атомарно, не забыть.

CREATE EXTENSION IF NOT EXISTS unaccent;

-- Кастомный конфиг russian_unaccent: russian + unaccent filter (для ё/е).
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_ts_config WHERE cfgname = 'russian_unaccent'
    ) THEN
        CREATE TEXT SEARCH CONFIGURATION russian_unaccent (COPY = russian);
        ALTER TEXT SEARCH CONFIGURATION russian_unaccent
            ALTER MAPPING FOR hword, hword_part, word
            WITH unaccent, russian_stem;
    END IF;
END$$;

ALTER TABLE prompts ADD COLUMN IF NOT EXISTS search_tsv tsvector
    GENERATED ALWAYS AS (
        setweight(to_tsvector('russian_unaccent', coalesce(title, '')), 'A') ||
        setweight(to_tsvector('english',          coalesce(title, '')), 'A') ||
        setweight(to_tsvector('russian_unaccent', coalesce(content, '')), 'B') ||
        setweight(to_tsvector('english',          coalesce(content, '')), 'B')
    ) STORED;

CREATE INDEX IF NOT EXISTS idx_prompts_search_tsv
    ON prompts USING GIN (search_tsv)
    WHERE deleted_at IS NULL;

ALTER TABLE prompts ALTER COLUMN search_tsv SET STATISTICS 1000;
