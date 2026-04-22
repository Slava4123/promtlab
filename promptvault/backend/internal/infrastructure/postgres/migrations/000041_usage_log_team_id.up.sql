-- prompt_usage_log.team_id — денормализация team_id из prompts для
-- быстрой team-аналитики без JOIN (dashboard-запросы heavy на JOIN
-- при миллионах записей в usage-логе).
--
-- Семантика: снапшот на момент use. Если промпт позже переносят в другую
-- команду — история использований остаётся в контексте «команды на момент
-- использования». Это корректно для аналитики (track past activity), а не
-- для текущей принадлежности.
--
-- Также заменяем idx_usage_log_prompt на compound (prompt_id, used_at DESC) —
-- prefix-совместимо со старым (фильтр по prompt_id продолжит работать),
-- но даёт ускорение per-prompt timeline запросам в /analytics/prompts/:id.

ALTER TABLE prompt_usage_log
    ADD COLUMN IF NOT EXISTS team_id BIGINT REFERENCES teams(id) ON DELETE SET NULL;

-- Частичный индекс: team_id чаще всего NULL (личные промпты), не раздуваем.
CREATE INDEX IF NOT EXISTS idx_pul_team_used
    ON prompt_usage_log (team_id, used_at DESC)
    WHERE team_id IS NOT NULL;

-- Замена индекса под per-prompt timeline. Safe: новый покрывает старый prefix.
DROP INDEX IF EXISTS idx_usage_log_prompt;
CREATE INDEX IF NOT EXISTS idx_usage_log_prompt_time
    ON prompt_usage_log (prompt_id, used_at DESC);

-- Backfill: team_id из текущего владения промпта. Safe: UPDATE только NULL-ы.
UPDATE prompt_usage_log pul
SET team_id = p.team_id
FROM prompts p
WHERE pul.prompt_id = p.id
  AND p.team_id IS NOT NULL
  AND pul.team_id IS NULL;
