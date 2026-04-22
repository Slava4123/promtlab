-- Reversible: восстанавливает схему с дефолтами, но исходные данные (default_model
-- пользователей, квоты планов) не вернёт — их нужно пересеять через seed при необходимости.

ALTER TABLE users ADD COLUMN IF NOT EXISTS default_model VARCHAR(100) DEFAULT 'anthropic/claude-sonnet-4';
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS max_ai_requests_daily INT NOT NULL DEFAULT 0;
ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS ai_requests_is_total BOOLEAN NOT NULL DEFAULT FALSE;
