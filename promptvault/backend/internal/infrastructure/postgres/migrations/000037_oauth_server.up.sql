-- OAuth 2.1 Authorization Server для внешних MCP-клиентов (claude.ai, Cursor и др.).
-- Согласно MCP spec 2025-06-18 §Authorization — OAuth 2.1 MUST + RFC 9728 + RFC 8414 +
-- RFC 8707 (Resource Indicators) + PKCE S256. Статические API-ключи pvlt_* остаются
-- без изменений в api_keys; эти три таблицы — параллельная авторизационная подсистема.
--
-- Scope-модель (read_only / allowed_tools / team_id) переиспользуется через jsonb-поле
-- `policy` — тот же models.Policy что и в apikey.KeyPolicy (см. internal/models/policy.go).

-- OAuth-клиенты: приложения, зарегистрировавшиеся через RFC 7591 Dynamic Client
-- Registration (Claude.ai, сторонние MCP-клиенты).
CREATE TABLE IF NOT EXISTS oauth_clients (
    id                           BIGSERIAL PRIMARY KEY,
    client_id                    TEXT        NOT NULL UNIQUE,
    client_secret_hash           TEXT,                           -- NULL = public client (PKCE-only)
    client_name                  TEXT        NOT NULL,
    redirect_uris                TEXT[]      NOT NULL,           -- whitelist callbacks
    grant_types                  TEXT[]      NOT NULL
        DEFAULT ARRAY['authorization_code', 'refresh_token'],
    response_types               TEXT[]      NOT NULL DEFAULT ARRAY['code'],
    token_endpoint_auth_method   TEXT        NOT NULL DEFAULT 'none',  -- none | client_secret_post
    scope                        TEXT        NOT NULL DEFAULT 'mcp:read mcp:write',
    is_dynamic                   BOOLEAN     NOT NULL DEFAULT TRUE,    -- TRUE = RFC 7591
    last_used_at                 TIMESTAMPTZ,
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_oauth_clients_last_used
    ON oauth_clients (last_used_at) WHERE last_used_at IS NOT NULL;

-- Authorization codes: короткоживущие one-time tokens для PKCE-flow.
-- TTL 60 сек согласно RFC 6749 §4.1.2 рекомендации.
-- code_hash = SHA256(raw_code) — сырой код есть только у клиента, не в БД.
CREATE TABLE IF NOT EXISTS oauth_authorization_codes (
    code_hash                    TEXT        PRIMARY KEY,        -- SHA256 от сырого кода
    client_id                    TEXT        NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    user_id                      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redirect_uri                 TEXT        NOT NULL,            -- должен совпадать с token-request
    code_challenge               TEXT        NOT NULL,            -- PKCE S256 challenge
    code_challenge_method        TEXT        NOT NULL DEFAULT 'S256',
    scope                        TEXT        NOT NULL,
    resource                     TEXT        NOT NULL,            -- RFC 8707 — целевой MCP URI
    policy                       JSONB       NOT NULL DEFAULT '{}'::jsonb,
    expires_at                   TIMESTAMPTZ NOT NULL,
    used_at                      TIMESTAMPTZ,                     -- one-time: NOT NULL после exchange
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_oauth_codes_expires ON oauth_authorization_codes (expires_at);
CREATE INDEX IF NOT EXISTS idx_oauth_codes_user    ON oauth_authorization_codes (user_id);

-- Access & refresh tokens:
--  access  — JWT RS256, БД-запись только для revoke-check. jti как token_hash.
--  refresh — opaque rf_*, SHA256 хэш.
-- Refresh-rotation (RFC 6749 §1.5 recommendation): новый refresh создаётся при каждом
-- token-exchange, старый помечается revoked_at. parent_token_id строит цепочку.
CREATE TABLE IF NOT EXISTS oauth_tokens (
    id                           BIGSERIAL PRIMARY KEY,
    token_type                   TEXT        NOT NULL CHECK (token_type IN ('access', 'refresh')),
    token_hash                   TEXT        NOT NULL UNIQUE,
    client_id                    TEXT        NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    user_id                      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope                        TEXT        NOT NULL,
    resource                     TEXT        NOT NULL,            -- RFC 8707 audience
    policy                       JSONB       NOT NULL DEFAULT '{}'::jsonb,
    expires_at                   TIMESTAMPTZ NOT NULL,
    revoked_at                   TIMESTAMPTZ,
    parent_token_id              BIGINT      REFERENCES oauth_tokens(id) ON DELETE SET NULL,
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_oauth_tokens_user_type ON oauth_tokens (user_id, token_type);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_client    ON oauth_tokens (client_id);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_expires   ON oauth_tokens (expires_at);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_parent    ON oauth_tokens (parent_token_id) WHERE parent_token_id IS NOT NULL;
