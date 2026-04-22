-- Phase 14 M10: email-уведомления по изменениям Smart Insights.
-- Одна строка на каждое отправленное уведомление (append-only).
-- Rate-limit — 1 письмо/неделю на пару (user_id, insight_type); enforce
-- на уровне insights_notifier через SELECT ... WHERE sent_at > NOW() - INTERVAL '7 days'.
CREATE TABLE IF NOT EXISTS insight_notifications (
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    insight_type TEXT   NOT NULL,
    sent_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, insight_type, sent_at)
);

CREATE INDEX IF NOT EXISTS idx_insight_notif_user_type_time
    ON insight_notifications(user_id, insight_type, sent_at DESC);

-- Opt-in toggle: users.insight_emails_enabled. Default false согласно
-- ФЗ-152 (требуется явное согласие на маркетинговую переписку).
ALTER TABLE users ADD COLUMN IF NOT EXISTS insight_emails_enabled BOOLEAN NOT NULL DEFAULT FALSE;
