-- Rollback: убираем 'paused' из разрешённых статусов. ВНИМАНИЕ: если в БД
-- есть строки со status='paused', этот rollback упадёт — сначала
-- мигрируйте такие записи в active/expired через приложение.

ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_status_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_status_check
    CHECK (status IN ('active', 'past_due', 'cancelled', 'expired'));
