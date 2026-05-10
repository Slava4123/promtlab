-- Восстанавливает прод-лимиты на free/pro/max после тестового низколимитного seed
-- (см. seed-test-data.sql). Числа взяты из миграций 000046_concrete_plan_limits,
-- 000053_prompt_chains и комментариев quota-exceeded-dialog.tsx.
--
-- Применение:
--   docker compose -f docker-compose.dev.yml exec -T postgres \
--     psql -U postgres -d promptvault -f /scripts/restore-prod-limits.sql

UPDATE subscription_plans SET
    max_prompts = CASE id
        WHEN 'free' THEN 50
        WHEN 'pro'  THEN 500
        WHEN 'pro_yearly' THEN 500
        WHEN 'max'  THEN 10000
        WHEN 'max_yearly' THEN 10000
    END,
    max_collections = CASE id
        WHEN 'free' THEN 3
        WHEN 'pro'  THEN 100
        WHEN 'pro_yearly' THEN 100
        WHEN 'max'  THEN 1000
        WHEN 'max_yearly' THEN 1000
    END,
    max_teams = CASE id
        WHEN 'free' THEN 1
        WHEN 'pro'  THEN 5
        WHEN 'pro_yearly' THEN 5
        WHEN 'max'  THEN 50
        WHEN 'max_yearly' THEN 50
    END,
    max_team_members = CASE id
        WHEN 'free' THEN 3
        WHEN 'pro'  THEN 10
        WHEN 'pro_yearly' THEN 10
        WHEN 'max'  THEN 50
        WHEN 'max_yearly' THEN 50
    END,
    max_share_links = CASE id
        WHEN 'free' THEN 2
        WHEN 'pro'  THEN 50
        WHEN 'pro_yearly' THEN 50
        WHEN 'max'  THEN 500
        WHEN 'max_yearly' THEN 500
    END,
    max_daily_shares = CASE id
        WHEN 'free' THEN 10
        WHEN 'pro'  THEN 100
        WHEN 'pro_yearly' THEN 100
        WHEN 'max'  THEN 1000
        WHEN 'max_yearly' THEN 1000
    END,
    max_ext_uses_daily = CASE id
        WHEN 'free' THEN 5
        WHEN 'pro'  THEN 100
        WHEN 'pro_yearly' THEN 100
        WHEN 'max'  THEN 500
        WHEN 'max_yearly' THEN 500
    END,
    max_mcp_uses_daily = CASE id
        WHEN 'free' THEN 5
        WHEN 'pro'  THEN 100
        WHEN 'pro_yearly' THEN 100
        WHEN 'max'  THEN 500
        WHEN 'max_yearly' THEN 500
    END,
    max_chains = CASE id
        WHEN 'free' THEN 1
        WHEN 'pro'  THEN 5
        WHEN 'pro_yearly' THEN 5
        WHEN 'max'  THEN 100
        WHEN 'max_yearly' THEN 100
    END,
    max_steps_per_chain = CASE id
        WHEN 'free' THEN 3
        WHEN 'pro'  THEN 10
        WHEN 'pro_yearly' THEN 10
        WHEN 'max'  THEN 50
        WHEN 'max_yearly' THEN 50
    END,
    max_saved_executions = CASE id
        WHEN 'free' THEN 0
        WHEN 'pro'  THEN 10
        WHEN 'pro_yearly' THEN 10
        WHEN 'max'  THEN 1000
        WHEN 'max_yearly' THEN 1000
    END,
    updated_at = now()
WHERE id IN ('free', 'pro', 'pro_yearly', 'max', 'max_yearly');

-- Подчищаем daily_feature_usage у тестовых юзеров (там могли быть backdate'ы и счётчики
-- от тестовых прогонов с низкими лимитами).
DELETE FROM daily_feature_usage
WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'e2e-%@test.local');

-- Verify
SELECT id, max_prompts, max_collections, max_teams, max_share_links, max_daily_shares,
       max_chains, max_steps_per_chain, max_saved_executions
FROM subscription_plans WHERE id IN ('free','pro','max') ORDER BY sort_order;
