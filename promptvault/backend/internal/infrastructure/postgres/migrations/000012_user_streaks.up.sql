CREATE TABLE IF NOT EXISTS user_streaks (
    user_id          BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current_streak   INTEGER NOT NULL DEFAULT 0,
    longest_streak   INTEGER NOT NULL DEFAULT 0,
    last_active_date DATE NOT NULL DEFAULT CURRENT_DATE,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
