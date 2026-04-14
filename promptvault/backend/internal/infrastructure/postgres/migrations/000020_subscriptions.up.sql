CREATE TABLE IF NOT EXISTS subscriptions (
    id                   BIGSERIAL PRIMARY KEY,
    user_id              BIGINT NOT NULL REFERENCES users(id),
    plan_id              VARCHAR(20) NOT NULL REFERENCES subscription_plans(id),
    status               VARCHAR(20) NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active','past_due','cancelled','expired')),
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end   TIMESTAMPTZ NOT NULL,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    cancelled_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_subscriptions_user_active
    ON subscriptions (user_id) WHERE status IN ('active','past_due');

CREATE INDEX idx_subscriptions_expiring
    ON subscriptions (status, current_period_end);
