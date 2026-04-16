-- M-6: пауза подписки на 1-3 месяца с auto-resume.
-- paused_at — момент входа в pause. Нужен для расчёта remaining при Resume:
--   remaining = current_period_end - paused_at; new_end = resume_now + remaining.
-- paused_until — запланированная дата авто-возобновления. ExpirationLoop
-- резюмит подписку при paused_until < now().
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS paused_at    TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS paused_until TIMESTAMPTZ;

-- Расширяем partial unique — у юзера должна быть только одна "живая" подписка,
-- включая paused. Иначе пользователь в pause мог бы завести вторую подписку.
DROP INDEX IF EXISTS idx_subscriptions_user_active;
CREATE UNIQUE INDEX idx_subscriptions_user_active
    ON subscriptions (user_id) WHERE status IN ('active','past_due','paused');

-- Индекс для быстрого поиска подписок, у которых пауза истекла — ExpirationLoop
-- сканит эту таблицу каждый тик.
CREATE INDEX IF NOT EXISTS idx_subscriptions_paused_resume
    ON subscriptions (paused_until) WHERE status = 'paused' AND paused_until IS NOT NULL;
