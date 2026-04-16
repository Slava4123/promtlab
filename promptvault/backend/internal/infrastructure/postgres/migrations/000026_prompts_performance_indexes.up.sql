-- P-1: композитный индекс для infinite-scroll списка промптов.
-- Текущий список фильтрует по user_id + team_id IS NULL + deleted_at IS NULL,
-- сортирует по updated_at DESC. Одиночный btree на user_id из 000001 не даёт
-- ordered-scan для ORDER BY updated_at — PG делает sort в памяти на каждый page.
-- Частичный индекс (WHERE deleted_at IS NULL) не хранит soft-удалённые строки,
-- что даёт лучшую заполненность листьев.
CREATE INDEX IF NOT EXISTS idx_prompts_user_updated
    ON prompts (user_id, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_prompts_team_updated
    ON prompts (team_id, updated_at DESC)
    WHERE team_id IS NOT NULL AND deleted_at IS NULL;

-- P-3: pg_trgm для ILIKE '%q%' поиска по title/content.
-- Без trgm-GIN индекса leading '%' в ILIKE делает полный sequential scan таблицы
-- на каждый keystroke (с дебаунсом фронта это всё равно десятки запросов в минуту
-- на 10k+ строк, блокирует connection pool).
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_prompts_title_trgm
    ON prompts USING gin (title gin_trgm_ops)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_prompts_content_trgm
    ON prompts USING gin (content gin_trgm_ops)
    WHERE deleted_at IS NULL;
