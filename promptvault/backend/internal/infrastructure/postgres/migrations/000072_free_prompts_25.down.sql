-- Откат Pack G: возвращаем Free max_prompts 25 → 15.
-- Grandfather через legacy_quotas (миграция 000068) остаётся нетронутым;
-- юзеры с legacy={max_prompts:50} продолжат видеть 50 (effectiveLimit max()).

UPDATE subscription_plans
   SET max_prompts = 15, updated_at = NOW()
 WHERE id = 'free';
