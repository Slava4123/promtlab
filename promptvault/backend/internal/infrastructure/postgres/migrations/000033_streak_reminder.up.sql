-- M-16: трекинг отправленных "не сломай серию" напоминаний, чтобы не слать
-- юзеру несколько раз в день при повторных тиках StreakReminderLoop.
ALTER TABLE user_streaks
    ADD COLUMN IF NOT EXISTS reminder_sent_on DATE;

-- Индекс для быстрого ListAtRisk — фильтр по current_streak + last_active_date.
CREATE INDEX IF NOT EXISTS idx_user_streaks_at_risk
    ON user_streaks (last_active_date) WHERE current_streak > 3;
