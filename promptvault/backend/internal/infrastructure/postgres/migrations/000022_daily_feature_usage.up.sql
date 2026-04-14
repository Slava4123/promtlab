CREATE TABLE IF NOT EXISTS daily_feature_usage (
    user_id      BIGINT NOT NULL REFERENCES users(id),
    usage_date   DATE NOT NULL,
    feature_type VARCHAR(20) NOT NULL,
    count        INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, usage_date, feature_type)
);
