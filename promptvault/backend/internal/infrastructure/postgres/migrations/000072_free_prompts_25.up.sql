-- Pack G: повышение Free max_prompts с 15 → 25.
-- Grandfather от Pack E (миграция 000068) сохраняется автоматически:
-- effectiveLimit = max(legacy_quotas.max_prompts, plan.MaxPrompts).
-- Юзеры с legacy={max_prompts:50} остаются на 50; новые юзеры — 25.

UPDATE subscription_plans
   SET max_prompts = 25, updated_at = NOW()
 WHERE id = 'free';
