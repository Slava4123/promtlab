-- Scoped API-keys: права (read_only), ограничение по команде, white-list tools, expiration.
-- Существующие ключи получают read_only=FALSE и NULL в остальных полях —
-- это означает полный доступ ко всем tools и командам без expiration (backward-compat).
-- team_id → ON DELETE CASCADE: если команда удалена, ключ для неё бесполезен.
-- CASCADE (а не SET NULL) предотвращает privilege escalation:
-- при SET NULL ограниченный ключ (team_id=42) превратился бы в unscoped ключ
-- (nil policy → полный доступ к личному пространству владельца).
ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS read_only     BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS team_id       BIGINT      REFERENCES teams(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS allowed_tools TEXT[],
    ADD COLUMN IF NOT EXISTS expires_at    TIMESTAMPTZ;

-- Частичный индекс для будущей очистки истёкших ключей.
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at
    ON api_keys (expires_at) WHERE expires_at IS NOT NULL;
