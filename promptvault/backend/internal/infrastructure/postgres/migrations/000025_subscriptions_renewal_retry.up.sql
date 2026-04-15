-- Retry-логика автопродления. Когда Charge возвращает ошибку (недостаточно средств,
-- карта истекла, банк-эмитент отклонил), подписка переводится в past_due и renewal loop
-- пытается списать повторно. last_renewal_attempt_at дросселирует повторные попытки
-- (не чаще чем раз в 24ч); renewal_attempts ограничивает общее число попыток (3)
-- перед финальным expired → free.
ALTER TABLE subscriptions ADD COLUMN last_renewal_attempt_at TIMESTAMPTZ;
ALTER TABLE subscriptions ADD COLUMN renewal_attempts        INTEGER NOT NULL DEFAULT 0;

-- Индекс для выборки past_due подписок, готовых к retry. Условие «last_attempt старше 24ч»
-- проверяется в WHERE запроса, но индекс по (status, last_renewal_attempt_at, attempts)
-- ускоряет полное сканирование.
CREATE INDEX idx_subscriptions_past_due_retry
    ON subscriptions (status, last_renewal_attempt_at, renewal_attempts)
    WHERE status = 'past_due';
