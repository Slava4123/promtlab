-- 000074_referral_pending_rewards.up.sql
-- Pricing iteration v3 (ADR-0009): delayed referral reward.
-- На webhook payment.succeeded → INSERT с eligible_at = now() + 14 дней.
-- ReferralRewardLoop ежечасно: SELECT WHERE eligible_at < now → grant + DELETE.
--
-- UNIQUE на referee_id — защита от double-INSERT при retry webhook'ов.
-- Один рефери → одна награда → одна запись в pending (до grant'а).

CREATE TABLE IF NOT EXISTS referral_pending_rewards (
    id          BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    referee_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    payment_id  BIGINT NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    eligible_at TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_referral_pending_unique_referee
    ON referral_pending_rewards (referee_id);

CREATE INDEX IF NOT EXISTS idx_referral_pending_eligible_at
    ON referral_pending_rewards (eligible_at);
