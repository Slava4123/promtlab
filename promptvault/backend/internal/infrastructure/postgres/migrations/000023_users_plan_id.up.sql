ALTER TABLE users ADD COLUMN plan_id VARCHAR(20) NOT NULL DEFAULT 'free'
    REFERENCES subscription_plans(id);

CREATE INDEX idx_users_plan ON users (plan_id);
