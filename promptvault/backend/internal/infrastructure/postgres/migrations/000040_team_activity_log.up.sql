-- team_activity_log — append-only лента продуктовых событий внутри команды
-- (Phase 14, задача 1 — кто/когда/что менял в промптах и других ресурсах).
--
-- Отличие от audit_log (000018): audit_log фиксирует админ-действия
-- (freeze_user, change_tier и т.д.) и полностью иммутабелен. Team activity
-- фиксирует продуктовые события (prompt.*, collection.*, share.*, member.*,
-- role.*) и видна всем членам команды (viewer+).
--
-- DENORMALIZATION: actor_email, actor_name, target_label хранятся снапшотом
-- (не FK). Это следует best practice для activity feed'ов:
--   1) запись переживает удаление user-а и target-ресурса (dead reference
--      без JOIN'а не ломает рендер);
--   2) feed рендерится без JOIN'ов — быстрее на больших командах.
-- Приватность: при удалении user-а запускается anonymize job
--   UPDATE team_activity_log SET actor_id=NULL, actor_email='deleted@anonymized',
--                                actor_name='(deleted user)' WHERE actor_id=?
-- (см. usecases/user/service.go после добавления в A.5).
--
-- APPEND-ONLY: триггер только на UPDATE. DELETE намеренно РАЗРЕШЁН —
-- нужен retention cleanup cron (Pro=90 дней, Max=365 дней). Tamper-evidence
-- держится на UPDATE-триггере: изменить существующую запись нельзя.
-- Для full-immutability (если потребуется) — переход на PARTITION BY RANGE
-- (created_at) + DETACH PARTITION для cleanup без DELETE.

CREATE TABLE IF NOT EXISTS team_activity_log (
    id           BIGSERIAL   PRIMARY KEY,
    team_id      BIGINT      NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    actor_id     BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    actor_email  VARCHAR(255) NOT NULL,
    actor_name   VARCHAR(255),
    event_type   VARCHAR(50) NOT NULL,
    target_type  VARCHAR(50) NOT NULL,
    target_id    BIGINT,
    target_label VARCHAR(500),
    metadata     JSONB,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Основной query — feed по команде с курсор-пагинацией:
--   SELECT ... WHERE team_id = ? AND created_at < cursor ORDER BY created_at DESC
CREATE INDEX IF NOT EXISTS idx_tal_team_created
    ON team_activity_log (team_id, created_at DESC);

-- Для склейки prompt_versions + team_activity_log на странице /prompts/:id/history
-- (частичный: target_id часто NULL после cascade delete).
CREATE INDEX IF NOT EXISTS idx_tal_target
    ON team_activity_log (target_type, target_id, created_at DESC)
    WHERE target_id IS NOT NULL;

-- Фильтр «события от конкретного автора» в UI feed.
CREATE INDEX IF NOT EXISTS idx_tal_team_actor
    ON team_activity_log (team_id, actor_id, created_at DESC)
    WHERE actor_id IS NOT NULL;

-- Tamper-evidence: ROW-level триггер разрешает UPDATE только actor_*
-- полей (id, team_id, event_type, target_*, metadata, created_at — иммутабельны).
--
-- Зачем row-level, а не STATEMENT как в audit_log (000018):
--   1) anonymize actor при удалении user (GDPR) — UPDATE actor_id/email/name
--   2) FK cascade ON DELETE SET NULL для actor_id
-- Оба сценария — это UPDATE'ы, которые должны пройти. STATEMENT-триггер
-- не даёт доступа к OLD/NEW, не может различить. Row-level решает.
-- Performance: activity_log редко получает UPDATE (anonymize один раз
-- в жизни юзера, cascade — тоже), row-level overhead незаметен.
CREATE OR REPLACE FUNCTION prevent_team_activity_log_mutation() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.id           IS DISTINCT FROM OLD.id
    OR NEW.team_id      IS DISTINCT FROM OLD.team_id
    OR NEW.event_type   IS DISTINCT FROM OLD.event_type
    OR NEW.target_type  IS DISTINCT FROM OLD.target_type
    OR NEW.target_id    IS DISTINCT FROM OLD.target_id
    OR NEW.target_label IS DISTINCT FROM OLD.target_label
    OR NEW.metadata     IS DISTINCT FROM OLD.metadata
    OR NEW.created_at   IS DISTINCT FROM OLD.created_at
    THEN
        RAISE EXCEPTION 'team_activity_log is append-only: only actor_* fields may be changed (for anonymize/FK cascade)'
            USING ERRCODE = 'insufficient_privilege';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS team_activity_log_prevent_mutation ON team_activity_log;

CREATE TRIGGER team_activity_log_prevent_mutation
    BEFORE UPDATE ON team_activity_log
    FOR EACH ROW
    EXECUTE FUNCTION prevent_team_activity_log_mutation();
