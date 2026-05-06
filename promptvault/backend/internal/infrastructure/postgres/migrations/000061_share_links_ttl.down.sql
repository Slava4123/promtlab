-- Откат: возвращаем колонки квот с дефолтами, восстанавливаем значения по plan_id.
-- expires_at оставляем — backfill необратим без historic знания, какие были NULL.

ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS max_share_links  INTEGER NOT NULL DEFAULT 2,
    ADD COLUMN IF NOT EXISTS max_daily_shares INTEGER NOT NULL DEFAULT 10;

UPDATE subscription_plans SET max_share_links = 50,  max_daily_shares = 100  WHERE id LIKE 'pro%';
UPDATE subscription_plans SET max_share_links = 500, max_daily_shares = 1000 WHERE id LIKE 'max%';
UPDATE subscription_plans SET max_share_links = 2,   max_daily_shares = 10   WHERE id = 'free';
