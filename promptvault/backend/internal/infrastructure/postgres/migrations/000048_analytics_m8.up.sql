-- Phase 14 M8: Full Smart Insights.
-- pg_trgm даёт similarity() / gin_trgm_ops для possible_duplicates инсайта.
-- IF NOT EXISTS — идемпотентность (managed PG может иметь extension предустановленным).
-- На managed Postgres (Timeweb/Яндекс) pg_trgm в pg_available_extensions;
-- если недоступен — миграция фэйлится, и feature-flag experimentalInsights
-- остаётся false в prod.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- GIN-индекс по content для быстрого similarity() с partial filter на not-deleted.
-- Partial index экономит место (удалённые промпты не индексируются).
CREATE INDEX IF NOT EXISTS idx_prompts_content_trgm
    ON prompts USING gin (content gin_trgm_ops)
    WHERE deleted_at IS NULL;
