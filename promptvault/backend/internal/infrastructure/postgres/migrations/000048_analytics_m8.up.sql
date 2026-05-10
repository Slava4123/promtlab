-- Phase 14 M8: Full Smart Insights.
-- pg_trgm даёт similarity() / gin_trgm_ops для possible_duplicates инсайта.
-- IF NOT EXISTS — идемпотентность (managed PG может иметь extension предустановленным).
-- На managed Postgres (Timeweb/Яндекс) pg_trgm в pg_available_extensions;
-- если недоступен — миграция фэйлится, и feature-flag experimentalInsights
-- остаётся false в prod.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- MN-81: idx_prompts_content_trgm уже создан в 000026_prompts_performance_indexes.up.sql,
-- здесь был второй CREATE INDEX IF NOT EXISTS — фактически no-op (IF NOT EXISTS),
-- но мусор: дубликат шумит при `pg_dump --schema-only` diff'ах и вводит в заблуждение
-- разработчиков (кажется что Phase 14 M8 владеет индексом). Оставляем только pg_trgm
-- extension; индекс остаётся в 000026 как single source of truth.
