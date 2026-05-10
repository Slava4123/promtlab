-- Откат: возвращаем Pro лимиты на 100/день. Удаляем legacy-ключи
-- max_ext_uses_daily и max_mcp_uses_daily (но оставляем max_prompts от
-- Pack E если есть — каждая миграция чистит только свои ключи).

UPDATE users
   SET legacy_quotas = legacy_quotas - 'max_ext_uses_daily' - 'max_mcp_uses_daily'
 WHERE plan_id IN ('pro', 'pro_yearly');

UPDATE subscription_plans
   SET max_ext_uses_daily = 100,
       max_mcp_uses_daily = 100,
       updated_at = NOW()
 WHERE id IN ('pro', 'pro_yearly');
