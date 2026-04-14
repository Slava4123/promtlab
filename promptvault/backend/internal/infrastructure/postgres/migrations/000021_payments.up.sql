CREATE TABLE IF NOT EXISTS payments (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    subscription_id BIGINT REFERENCES subscriptions(id),
    external_id     VARCHAR(100) NOT NULL,
    idempotency_key VARCHAR(100) NOT NULL UNIQUE,
    amount_kop      INTEGER NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'RUB',
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','succeeded','failed','refunded')),
    provider        VARCHAR(20) NOT NULL DEFAULT 'tbank',
    provider_data   JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_payments_external ON payments (provider, external_id);
CREATE INDEX idx_payments_user ON payments (user_id, created_at DESC);
