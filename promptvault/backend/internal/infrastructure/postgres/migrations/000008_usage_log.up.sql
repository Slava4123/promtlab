CREATE TABLE IF NOT EXISTS prompt_usage_log (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    prompt_id  BIGINT NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    used_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_log_user_time ON prompt_usage_log (user_id, used_at DESC);
CREATE INDEX idx_usage_log_prompt ON prompt_usage_log (prompt_id);
