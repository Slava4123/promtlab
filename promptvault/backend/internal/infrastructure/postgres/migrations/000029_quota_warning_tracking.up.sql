-- Трекинг последнего отправленного quota-warning email, чтобы не слать юзеру
-- уведомление несколько раз в день при каждом новом AI-запросе >= 80% лимита.
-- DATE (не TIMESTAMPTZ): сверяемся с today в user-tz через X-Timezone header
-- при IncrementAIUsage; нам важна только дата.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS quota_warning_sent_on DATE;
