ALTER TABLE users
    ADD COLUMN IF NOT EXISTS onboarding_completed_at TIMESTAMPTZ NULL;

-- Backfill: existing users считаются уже прошедшими онбординг,
-- чтобы не показывать им wizard принудительно после деплоя.
UPDATE users SET onboarding_completed_at = NOW() WHERE onboarding_completed_at IS NULL;
