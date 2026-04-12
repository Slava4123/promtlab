-- TOTP 2FA для админов. Один secret на юзера (PRIMARY KEY user_id),
-- confirmed_at=NULL означает что enrollment начат но ещё не подтверждён
-- первым кодом из Authenticator (unconfirmed enrollment можно безопасно
-- перезаписать при re-enroll).

CREATE TABLE IF NOT EXISTS user_totp (
    user_id      BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret       VARCHAR(64) NOT NULL,
    confirmed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Backup codes: 10 одноразовых кодов, хранятся как bcrypt-хеши (как passwords).
-- Когда юзер теряет телефон, вводит один backup code → used_at=NOW.
CREATE TABLE IF NOT EXISTS user_totp_backup_codes (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash  VARCHAR(128) NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Partial index: для быстрого listing неиспользованных кодов.
CREATE INDEX IF NOT EXISTS idx_user_totp_backup_codes_user ON user_totp_backup_codes (user_id) WHERE used_at IS NULL;
