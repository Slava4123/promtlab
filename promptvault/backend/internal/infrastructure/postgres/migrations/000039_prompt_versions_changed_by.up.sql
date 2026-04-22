-- prompt_versions.changed_by — автор версии для team activity feed (Phase 14).
--
-- До этой миграции версии хранили только снапшот (title/content/model/change_note)
-- без информации о том, КТО именно создал версию. В команде это критично.
--
-- nullable: старые версии заполняются как автор промпта (владелец). При удалении
-- user-а ON DELETE SET NULL — версия остаётся, автор обнуляется; UI отрисует
-- «(удалённый пользователь)».

ALTER TABLE prompt_versions
    ADD COLUMN IF NOT EXISTS changed_by BIGINT REFERENCES users(id) ON DELETE SET NULL;

-- Частичный индекс: уменьшает размер (NULL'ы не индексируются) и ускоряет
-- фильтр «все версии автора X» на странице /analytics/contributors.
CREATE INDEX IF NOT EXISTS idx_prompt_versions_changed_by
    ON prompt_versions (changed_by)
    WHERE changed_by IS NOT NULL;

-- Backfill: автор старых версий = владелец промпта. Safe: UPDATE только NULL-ы.
UPDATE prompt_versions pv
SET changed_by = p.user_id
FROM prompts p
WHERE pv.prompt_id = p.id
  AND pv.changed_by IS NULL;
