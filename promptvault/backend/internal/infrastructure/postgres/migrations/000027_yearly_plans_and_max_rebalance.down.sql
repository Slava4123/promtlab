DELETE FROM subscription_plans WHERE id IN ('pro_yearly', 'max_yearly');

UPDATE subscription_plans
   SET max_ai_requests_daily = 30,
       updated_at            = NOW()
 WHERE id = 'max';
