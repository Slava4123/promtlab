-- Max с 30 AI/день слишком щедр: worst-case маржа 24% при cost 1.1₽/запрос.
-- 15/день = 450/мес × 1.1₽ = 495₽ расход, маржа 62% на 1299₽ — устойчиво.
UPDATE subscription_plans
   SET max_ai_requests_daily = 15,
       updated_at            = NOW()
 WHERE id = 'max';

-- Yearly планы (−10% от месячной × 12). Меньшая скидка не даёт анкоринг-эффекта,
-- большая — съедает маржу ниже 35% на worst-case.
--   Pro monthly: 599 × 12 = 7188₽, yearly 6490₽ (−9.7%)
--   Max monthly: 1299 × 12 = 15588₽, yearly 13990₽ (−10.2%)
-- Лимиты идентичны monthly-аналогу. period_days=365.
INSERT INTO subscription_plans (id, name, price_kop, period_days,
    max_prompts, max_collections, max_ai_requests_daily, ai_requests_is_total,
    max_teams, max_team_members, max_share_links, max_ext_uses_daily,
    max_mcp_uses_daily, features, sort_order)
VALUES
    ('pro_yearly', 'Pro (год)', 649000, 365,
     500, -1, 10, FALSE,
     5, 10, 10, 30, 30,
     '["priority_support", "yearly"]'::jsonb, 3),
    ('max_yearly', 'Max (год)', 1399000, 365,
     -1, -1, 15, FALSE,
     -1, -1, -1, -1, -1,
     '["priority_support", "yearly"]'::jsonb, 4)
ON CONFLICT (id) DO NOTHING;
