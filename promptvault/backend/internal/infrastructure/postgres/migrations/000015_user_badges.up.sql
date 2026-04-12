CREATE TABLE IF NOT EXISTS user_badges (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_id    VARCHAR(50) NOT NULL,
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_badges_user_badge ON user_badges (user_id, badge_id);
CREATE INDEX IF NOT EXISTS idx_user_badges_user_id ON user_badges (user_id);
