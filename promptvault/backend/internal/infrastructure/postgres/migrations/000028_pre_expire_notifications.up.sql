-- Трекинг отправленных pre-expire напоминаний, чтобы не спамить юзера на каждом
-- тике ReminderLoop. Stage: 0 — не отправляли, 1 — отправили 3-day reminder,
-- 2 — отправили 1-day. Сбрасывается при продлении (ExtendPeriod).
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS pre_expire_stage SMALLINT NOT NULL DEFAULT 0;

-- Welcome email тоже трекается в users, чтобы повторный verify (edge case через
-- admin или ручное обновление) не отправил письмо второй раз.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS welcome_sent_at TIMESTAMPTZ;

-- Re-engagement email (M-5d) — трекаем дату последней отправки, чтобы не слать
-- чаще одного раза в 30 дней даже если юзер так и не вернулся.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS reengagement_sent_at TIMESTAMPTZ;

-- last_login_at для M-5d re-engagement. Обновляется в auth.Login и issueTokens.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;
