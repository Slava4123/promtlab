-- E2E test data: переписывает прод-планы free/pro/max низкими лимитами для dev-стека
-- + создаёт 3 тестовых юзера с активной подпиской.
-- Цель: Playwright/UI тесты упираются в лимит за 1-2 действия вместо 50-500.
--
-- Применение:
--   docker compose -f docker-compose.dev.yml exec -T postgres \
--     psql -U postgres -d promptvault -f /scripts/seed-test-data.sql
--
-- ⚠️  Этот скрипт МЕНЯЕТ ЛИМИТЫ В ОСНОВНЫХ ПЛАНАХ free/pro/max.
-- Запускать ТОЛЬКО в dev-окружении! В prod нельзя — обнулит реальные лимиты.
--
-- Идемпотентен: ON CONFLICT DO UPDATE.
-- Пароль для всех 3 юзеров: TestPass2026! (хэш через pgcrypto bcrypt cost=10).

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ===== Лимиты прод-планов под E2E =====
-- Числа подобраны так, чтобы спецы упирались в блок за 2-3 действия.
-- max_steps_per_chain не <2 (1-шаговая цепочка вырождена).
-- max_mcp_uses_daily не трогаем (MCP-тесты вне scope).
UPDATE subscription_plans SET
    max_prompts          = CASE id WHEN 'free' THEN 1 WHEN 'pro' THEN 2 WHEN 'max' THEN 3 END,
    max_collections      = CASE id WHEN 'free' THEN 1 WHEN 'pro' THEN 2 WHEN 'max' THEN 3 END,
    max_teams            = CASE id WHEN 'free' THEN 1 WHEN 'pro' THEN 2 WHEN 'max' THEN 3 END,
    max_team_members     = CASE id WHEN 'free' THEN 1 WHEN 'pro' THEN 2 WHEN 'max' THEN 3 END,
    max_share_links      = CASE id WHEN 'free' THEN 1 WHEN 'pro' THEN 2 WHEN 'max' THEN 3 END,
    max_daily_shares     = CASE id WHEN 'free' THEN 1 WHEN 'pro' THEN 2 WHEN 'max' THEN 3 END,
    max_ext_uses_daily   = CASE id WHEN 'free' THEN 1 WHEN 'pro' THEN 2 WHEN 'max' THEN 3 END,
    max_chains           = CASE id WHEN 'free' THEN 1 WHEN 'pro' THEN 2 WHEN 'max' THEN 3 END,
    max_steps_per_chain  = CASE id WHEN 'free' THEN 2 WHEN 'pro' THEN 3 WHEN 'max' THEN 4 END,
    max_saved_executions = CASE id WHEN 'free' THEN 0 WHEN 'pro' THEN 1 WHEN 'max' THEN 2 END,
    updated_at           = now()
WHERE id IN ('free', 'pro', 'max');

-- Подчищаем устаревшие test_-планы если они остались с предыдущих прогонов.
DELETE FROM subscription_plans WHERE id IN ('test_free', 'test_pro', 'test_max');

-- ===== Users =====
-- pgcrypto bcrypt совместим с golang.org/x/crypto/bcrypt — оба используют $2a$ формат.
-- referral_code NOT NULL UNIQUE — детерминированное значение per-tier.
-- onboarding_completed_at — НЕ NULL: иначе фронт редиректит на /welcome wizard.
INSERT INTO users (
    email, name, password_hash, plan_id, email_verified, role, status,
    referral_code, onboarding_completed_at, created_at, updated_at
) VALUES
    ('e2e-free@test.local', 'E2E Free User', crypt('TestPass2026!', gen_salt('bf', 10)), 'free', true, 'user', 'active', 'TESTREFFREE', now(), now(), now()),
    ('e2e-pro@test.local',  'E2E Pro User',  crypt('TestPass2026!', gen_salt('bf', 10)), 'pro',  true, 'user', 'active', 'TESTREFPRO',  now(), now(), now()),
    ('e2e-max@test.local',  'E2E Max User',  crypt('TestPass2026!', gen_salt('bf', 10)), 'max',  true, 'user', 'active', 'TESTREFMAX',  now(), now(), now())
ON CONFLICT (email) DO UPDATE SET
    plan_id                 = EXCLUDED.plan_id,
    email_verified          = true,
    status                  = 'active',
    onboarding_completed_at = COALESCE(users.onboarding_completed_at, now()),
    -- пароль обновляем на свежий хэш — на случай если ранее был старый
    password_hash           = EXCLUDED.password_hash,
    updated_at              = now();

-- Если у юзеров есть подписки от предыдущих прогонов с устаревшими test_-плану,
-- сразу обновляем их plan_id и продлеваем active до 30 дней.
UPDATE subscriptions
SET plan_id    = u.plan_id,
    status     = 'active',
    current_period_start = now(),
    current_period_end   = now() + interval '30 days',
    updated_at = now()
FROM users u
WHERE subscriptions.user_id = u.id
  AND u.email IN ('e2e-free@test.local', 'e2e-pro@test.local', 'e2e-max@test.local');

-- ===== Subscriptions =====
INSERT INTO subscriptions (
    user_id, plan_id, status, current_period_start, current_period_end,
    cancel_at_period_end, auto_renew, renewal_attempts, pre_expire_stage,
    created_at, updated_at
)
SELECT u.id, u.plan_id, 'active', now(), now() + interval '30 days', false, false, 0, 0, now(), now()
FROM users u
WHERE u.email IN ('e2e-free@test.local', 'e2e-pro@test.local', 'e2e-max@test.local')
ON CONFLICT DO NOTHING;

-- ===== Verify =====
SELECT u.email, u.plan_id, p.max_prompts, p.max_chains, p.max_steps_per_chain,
       s.status AS sub_status, s.current_period_end::timestamp(0)
FROM users u
JOIN subscription_plans p ON p.id = u.plan_id
JOIN subscriptions s ON s.user_id = u.id
WHERE u.email LIKE 'e2e-%@test.local'
ORDER BY p.sort_order;
