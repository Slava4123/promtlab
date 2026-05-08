-- migrate:no-transaction
-- MN-39: functional indexes для prompt.SuggestByPrefix.
-- Запрос делает `lower(title) LIKE 'prefix%'` (см. prompt_repo.go:230,232).
-- Без functional index PG делает full scan по таблице. Индекс с
-- `text_pattern_ops` оптимизирован под LIKE-запросы с anchored prefix
-- (без leading wildcard).
--
-- Два индекса — по (user_id, lower(title)) для personal-режима и
-- (team_id, lower(title)) для team-режима. Покрывают оба ветки в
-- SuggestByPrefix.
--
-- CREATE INDEX CONCURRENTLY (CR-16): без блокировки writes.

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_prompts_user_title_lower_prefix
    ON prompts (user_id, lower(title) text_pattern_ops)
    WHERE team_id IS NULL AND deleted_at IS NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_prompts_team_title_lower_prefix
    ON prompts (team_id, lower(title) text_pattern_ops)
    WHERE team_id IS NOT NULL AND deleted_at IS NULL;
