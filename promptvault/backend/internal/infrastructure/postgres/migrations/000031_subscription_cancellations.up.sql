-- M-6b: exit survey при отмене подписки.
-- Запись причины cancel для аналитики. Одна запись на факт отмены,
-- не update — если юзер передумал и Resume, а потом снова Cancel, это
-- две отдельные записи.
--
-- reason — свободная строка, но контролируется валидатором на Go-стороне
-- (too_expensive / not_using / missing_feature / found_alternative / other).
-- other_text — заполняется только при reason='other'.
CREATE TABLE IF NOT EXISTS subscription_cancellations (
    id              SERIAL PRIMARY KEY,
    user_id         BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id BIGINT      NOT NULL,
    plan_id         VARCHAR(20) NOT NULL,
    reason          VARCHAR(30) NOT NULL,
    other_text      TEXT        NOT NULL DEFAULT '',
    cancelled_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscription_cancellations_cancelled_at
    ON subscription_cancellations (cancelled_at DESC);

CREATE INDEX IF NOT EXISTS idx_subscription_cancellations_user
    ON subscription_cancellations (user_id, cancelled_at DESC);
