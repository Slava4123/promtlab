CREATE TABLE IF NOT EXISTS share_links (
    id             BIGSERIAL    PRIMARY KEY,
    prompt_id      BIGINT       NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    user_id        BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token          VARCHAR(64)  NOT NULL UNIQUE,
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    view_count     INTEGER      NOT NULL DEFAULT 0,
    last_viewed_at TIMESTAMPTZ,
    expires_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_share_links_active_prompt
    ON share_links (prompt_id) WHERE is_active = true;
CREATE INDEX idx_share_links_user ON share_links (user_id);
