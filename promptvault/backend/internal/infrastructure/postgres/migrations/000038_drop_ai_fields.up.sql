-- Drop AI-related fields after full removal of OpenRouter integration.
-- users.default_model — последняя выбранная модель для AI-ассистента.
-- subscription_plans.max_ai_requests_daily / ai_requests_is_total — AI-квоты по тарифу.
-- daily_feature_usage остаётся: feature_type='ai' перестаёт записываться, старые строки не трогаем.

ALTER TABLE users DROP COLUMN IF EXISTS default_model;
ALTER TABLE subscription_plans DROP COLUMN IF EXISTS max_ai_requests_daily;
ALTER TABLE subscription_plans DROP COLUMN IF EXISTS ai_requests_is_total;
