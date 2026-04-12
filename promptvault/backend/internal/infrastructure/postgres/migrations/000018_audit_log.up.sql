-- audit_log — append-only журнал административных действий для compliance
-- (OWASP A09:2025). before_state / after_state — JSONB с не-sensitive metadata
-- (ни password_hash, ни JWT, ни TOTP secrets не кладутся).

CREATE TABLE IF NOT EXISTS audit_log (
    id           BIGSERIAL PRIMARY KEY,
    admin_id     BIGINT NOT NULL REFERENCES users(id),
    action       VARCHAR(50) NOT NULL,
    target_type  VARCHAR(50) NOT NULL,
    target_id    BIGINT,
    before_state JSONB,
    after_state  JSONB,
    ip           INET NOT NULL,
    user_agent   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индексы под основные запросы админ-панели (audit log feed).
CREATE INDEX IF NOT EXISTS idx_audit_log_admin_created ON audit_log (admin_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_target ON audit_log (target_type, target_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_action_created ON audit_log (action, created_at DESC);

-- Append-only enforcement через BEFORE UPDATE/DELETE триггеры.
-- Выбран trigger-based подход, а НЕ REVOKE UPDATE, DELETE, потому что
-- REVOKE не действует на owner таблицы — а в большинстве single-role setup
-- (включая наш docker-compose) app подключается именно как owner. Trigger
-- срабатывает для ВСЕХ операций, включая от owner.
--
-- Чтобы обойти защиту, нужно вручную дропать триггеры в psql как superuser —
-- это явный intent, требующий separate admin session.
CREATE OR REPLACE FUNCTION prevent_audit_log_modification() RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_log is append-only: % operations are not allowed', TG_OP
        USING ERRCODE = 'insufficient_privilege';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS audit_log_prevent_update ON audit_log;
DROP TRIGGER IF EXISTS audit_log_prevent_delete ON audit_log;

CREATE TRIGGER audit_log_prevent_update
    BEFORE UPDATE ON audit_log
    FOR EACH STATEMENT
    EXECUTE FUNCTION prevent_audit_log_modification();

CREATE TRIGGER audit_log_prevent_delete
    BEFORE DELETE ON audit_log
    FOR EACH STATEMENT
    EXECUTE FUNCTION prevent_audit_log_modification();
