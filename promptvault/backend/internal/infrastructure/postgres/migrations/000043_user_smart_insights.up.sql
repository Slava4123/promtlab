-- user_smart_insights — кэш детерминированных инсайтов для Max-пользователей
-- (Phase 14, задача 2 — Smart Insights). Вычисляются ежедневным cron-job'ом
-- (usecases/analytics/insights.go) и показываются на /analytics → Insights panel.
--
-- insight_type: 'unused_prompts' | 'trending' | 'declining' | 'most_edited' |
--               'possible_duplicates' | 'orphan_tags' | 'empty_collections'
-- payload: JSONB-структура, зависит от insight_type. Примеры:
--   unused_prompts: [{"prompt_id": 42, "title": "...", "last_used_at": "..."}, ...]
--   trending:       [{"prompt_id": 42, "uses_last_7d": 120, "uses_prev_7d": 40}, ...]
--
-- team_id — nullable: инсайт может быть персональным (team_id=NULL) или
-- командным (team_id=X). UNIQUE по (user_id, team_id, insight_type) — один
-- актуальный набор на скоуп.

CREATE TABLE IF NOT EXISTS user_smart_insights (
    id           BIGSERIAL   PRIMARY KEY,
    user_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id      BIGINT      REFERENCES teams(id) ON DELETE CASCADE,
    insight_type VARCHAR(50) NOT NULL,
    payload      JSONB       NOT NULL,
    computed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- UNIQUE с COALESCE для NULL team_id: без COALESCE уникальность не сработает,
-- т.к. NULL != NULL в SQL.
CREATE UNIQUE INDEX IF NOT EXISTS idx_usi_unique
    ON user_smart_insights (user_id, COALESCE(team_id, 0), insight_type);

-- Для выборки всех инсайтов юзера на dashboard:
CREATE INDEX IF NOT EXISTS idx_usi_user_computed
    ON user_smart_insights (user_id, computed_at DESC);
