-- MN-81: 000048 больше не владеет idx_prompts_content_trgm — единственный
-- источник истины 000026. Здесь раньше был DROP INDEX, который при rollback
-- 000048 → 000047 удалял индекс, делая 000026 в incoherent state.
-- Extension не дропаем: могут быть другие индексы, зависящие от pg_trgm.
SELECT 1; -- no-op
