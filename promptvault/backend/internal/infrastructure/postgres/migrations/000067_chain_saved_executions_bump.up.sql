-- Phase 16 quotas polish: повышаем max_saved_executions для Free и Pro,
-- чтобы исправить UX-баг (Free тратил токены на свой LLM и история не
-- сохранялась, Pro упирался в 10 на 5-й запуск).
--
-- Числа калиброваны по best-practices freemium-моделей (Notion 7→30→∞ дней
-- истории, Stanford 80/3, behavioral trigger principle):
--   Free  3   — даёт реальную «попробовать историю» (видишь предыдущие
--               запуски, сравниваешь, делишься), упирается на 4-м запуске,
--               что соответствует Notion-style 7-day window.
--   Pro   50  — типичный workflow (5 цепочек × 10 запусков) проходит,
--               но активный power-user через 1-2 месяца упирается → strong
--               upsell к Max. Прежние 100 были soft-безлимитом без давления.
--   Max   1000 — без изменений, достаточно для агентской работы.

UPDATE subscription_plans
   SET max_saved_executions = 3, updated_at = NOW()
 WHERE id = 'free';

UPDATE subscription_plans
   SET max_saved_executions = 50, updated_at = NOW()
 WHERE id IN ('pro', 'pro_yearly');
