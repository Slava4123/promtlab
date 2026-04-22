-- subscription_plans.max_daily_shares — дневной лимит на СОЗДАНИЕ публичных
-- шар-ссылок (Phase 14, задача 3). Заменяет семантику «total active» на
-- «created per day» (fixed window, полночь UTC, через daily_feature_usage
-- feature_type='share_create').
--
-- Цифры согласованы с user:  Free=10, Pro=100, Max=1000 в день.
-- DEFAULT 10 — бережный fallback для любых будущих планов.
--
-- Существующее поле max_share_links оставляем КАК soft cap на одновременно
-- активные ссылки (опционально — можно отключить в usecase, передав -1).
-- Удалять его миграцией сейчас не будем — безопаснее поэтапно.

ALTER TABLE subscription_plans
    ADD COLUMN IF NOT EXISTS max_daily_shares INT NOT NULL DEFAULT 10;

-- Заполняем по тарифам. Префиксный LIKE покрывает годовые варианты
-- (pro, pro_yearly; max, max_yearly).
UPDATE subscription_plans SET max_daily_shares = 100  WHERE id LIKE 'pro%';
UPDATE subscription_plans SET max_daily_shares = 1000 WHERE id LIKE 'max%';
