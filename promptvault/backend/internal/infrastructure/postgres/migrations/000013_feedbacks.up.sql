CREATE TYPE feedback_type AS ENUM ('bug', 'feature', 'other');

CREATE TABLE IF NOT EXISTS feedbacks (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type       feedback_type NOT NULL DEFAULT 'other',
    message    TEXT NOT NULL,
    page_url   VARCHAR(2000) DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_feedbacks_user_id ON feedbacks (user_id);
CREATE INDEX IF NOT EXISTS idx_feedbacks_created_at ON feedbacks (created_at DESC);
