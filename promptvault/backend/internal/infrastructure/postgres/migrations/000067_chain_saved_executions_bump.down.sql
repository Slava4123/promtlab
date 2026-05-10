-- Откат к значениям из миграции 000053.
UPDATE subscription_plans
   SET max_saved_executions = 0, updated_at = NOW()
 WHERE id = 'free';

UPDATE subscription_plans
   SET max_saved_executions = 10, updated_at = NOW()
 WHERE id IN ('pro', 'pro_yearly');
