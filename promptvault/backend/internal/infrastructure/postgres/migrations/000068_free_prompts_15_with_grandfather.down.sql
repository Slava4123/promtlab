-- Откат: возвращаем Free max_prompts=50 и удаляем колонку legacy_quotas.
-- ВАЖНО: после применения этой миграции Pack F (000070+) может прекратить
-- работать корректно, если использовал legacy_quotas — down должен идти в
-- порядке F → E.

UPDATE subscription_plans
   SET max_prompts = 50, updated_at = NOW()
 WHERE id = 'free';

ALTER TABLE users DROP COLUMN IF EXISTS legacy_quotas;
