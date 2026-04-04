-- 000001_initial_schema.up.sql
-- Initial schema matching GORM models as of migration from AutoMigrate.

CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) DEFAULT '',
    name          VARCHAR(100) NOT NULL,
    username      VARCHAR(30) DEFAULT '',
    avatar_url    VARCHAR(500) DEFAULT '',
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    default_model VARCHAR(100) DEFAULT 'anthropic/claude-sonnet-4',
    token_nonce   VARCHAR(64) DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username) WHERE username != '';

CREATE TABLE IF NOT EXISTS linked_accounts (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    provider    VARCHAR(20) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_provider ON linked_accounts (user_id, provider);
CREATE INDEX IF NOT EXISTS idx_linked_accounts_provider_id ON linked_accounts (provider_id);

CREATE TABLE IF NOT EXISTS email_verifications (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL,
    code       VARCHAR(6) NOT NULL,
    attempts   INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_email_verifications_user_id ON email_verifications (user_id);

CREATE TABLE IF NOT EXISTS teams (
    id          BIGSERIAL PRIMARY KEY,
    slug        VARCHAR(100) NOT NULL,
    name        VARCHAR(200) NOT NULL,
    description VARCHAR(500) DEFAULT '',
    created_by  BIGINT NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_teams_slug ON teams (slug);

CREATE TABLE IF NOT EXISTS team_members (
    id      BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id),
    user_id BIGINT NOT NULL REFERENCES users(id),
    role    VARCHAR(20) NOT NULL DEFAULT 'viewer'
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_team_user ON team_members (team_id, user_id);

CREATE TABLE IF NOT EXISTS team_invitations (
    id         BIGSERIAL PRIMARY KEY,
    team_id    BIGINT NOT NULL REFERENCES teams(id),
    user_id    BIGINT NOT NULL REFERENCES users(id),
    inviter_id BIGINT NOT NULL REFERENCES users(id),
    role       VARCHAR(20) NOT NULL DEFAULT 'viewer',
    status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_team_invitations_team_id ON team_invitations (team_id);
CREATE INDEX IF NOT EXISTS idx_team_invitations_user_id ON team_invitations (user_id);

CREATE TABLE IF NOT EXISTS collections (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    team_id     BIGINT,
    name        VARCHAR(200) NOT NULL,
    description VARCHAR(500) DEFAULT '',
    color       VARCHAR(20) DEFAULT '#8b5cf6',
    icon        VARCHAR(10) DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_collections_user_id ON collections (user_id);

CREATE TABLE IF NOT EXISTS prompts (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    team_id     BIGINT,
    title       VARCHAR(300) NOT NULL,
    content     TEXT NOT NULL,
    model       VARCHAR(100) DEFAULT '',
    favorite    BOOLEAN NOT NULL DEFAULT FALSE,
    usage_count INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_prompts_user_id ON prompts (user_id);
CREATE INDEX IF NOT EXISTS idx_prompts_team_id ON prompts (team_id);
CREATE INDEX IF NOT EXISTS idx_prompts_deleted_at ON prompts (deleted_at);

CREATE TABLE IF NOT EXISTS tags (
    id      BIGSERIAL PRIMARY KEY,
    name    VARCHAR(50) NOT NULL,
    color   VARCHAR(7) DEFAULT '#6366f1',
    user_id BIGINT,
    team_id BIGINT
);
CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags (user_id);
CREATE INDEX IF NOT EXISTS idx_tags_team_id ON tags (team_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_name_user_team ON tags (name, user_id, COALESCE(team_id, 0));

CREATE TABLE IF NOT EXISTS prompt_versions (
    id             BIGSERIAL PRIMARY KEY,
    prompt_id      BIGINT NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    version_number BIGINT NOT NULL,
    title          VARCHAR(300) NOT NULL,
    content        TEXT NOT NULL,
    model          VARCHAR(100) DEFAULT '',
    change_note    VARCHAR(300) DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_prompt_version ON prompt_versions (prompt_id, version_number);

-- Join tables
CREATE TABLE IF NOT EXISTS prompt_tags (
    prompt_id BIGINT NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    tag_id    BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (prompt_id, tag_id)
);

CREATE TABLE IF NOT EXISTS prompt_collections (
    prompt_id     BIGINT NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    collection_id BIGINT NOT NULL REFERENCES collections(id) ON DELETE CASCADE,
    PRIMARY KEY (prompt_id, collection_id)
);
