-- Bugfix: миграция 000030_subscription_pause расширила индекс
-- idx_subscriptions_user_active чтобы включить status='paused', но забыла
-- обновить CHECK constraint subscriptions_status_check. В результате
-- ЛЮБОЙ POST /api/subscription/pause валился 500 — backend пытается
-- UPDATE status='paused', PG отвергает по CHECK.
--
-- Восстанавливаем consistent набор статусов: active, past_due, cancelled,
-- expired, paused.

-- MN-80: NOT VALID + VALIDATE через 2 шага вместо одного ADD CONSTRAINT.
--   ADD CONSTRAINT без NOT VALID берёт AccessExclusiveLock и сканирует всю
--   таблицу для проверки existing rows. На большой subscriptions таблице
--   это блокирует webhook-обработку (т.к. webhooks делают INSERT/UPDATE).
--
--   NOT VALID — берёт ShareUpdate lock на короткое время (только метаданные),
--   старые данные не проверяет (только новые INSERT/UPDATE). Затем VALIDATE
--   CONSTRAINT берёт ShareUpdate lock и сканирует таблицу — но уже не блокирует
--   write-операции. Если миграция применяется при наличии трафика, это разница
--   между downtime'ом 30+ секунд и near-zero.
--
--   На subscriptions сейчас < 1k rows, разницы не видно. Но pattern важен
--   как образец для следующих CHECK-миграций (см. docs/MIGRATIONS.md).

ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS subscriptions_status_check;
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_status_check
    CHECK (status IN ('active', 'past_due', 'cancelled', 'expired', 'paused'))
    NOT VALID;
ALTER TABLE subscriptions VALIDATE CONSTRAINT subscriptions_status_check;
