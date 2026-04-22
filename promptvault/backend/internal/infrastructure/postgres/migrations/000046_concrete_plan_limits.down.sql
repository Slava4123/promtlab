-- Откат к "безлимитному" представлению через sentinel-значение -1.
-- Соответствует исходному seed'у в 000019_subscription_plans.up.sql.

UPDATE subscription_plans SET
    max_collections = -1
WHERE id IN ('pro', 'pro_yearly');

UPDATE subscription_plans SET
    max_prompts        = -1,
    max_collections    = -1,
    max_teams          = -1,
    max_share_links    = -1,
    max_ext_uses_daily = -1,
    max_mcp_uses_daily = -1
WHERE id IN ('max', 'max_yearly');

-- Free — без изменений (там -1 никогда не было).
