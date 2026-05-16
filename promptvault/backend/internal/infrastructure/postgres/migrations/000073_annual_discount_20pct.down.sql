UPDATE subscription_plans
   SET price_kop = 649000, updated_at = NOW()
 WHERE id = 'pro_yearly';

UPDATE subscription_plans
   SET price_kop = 1399000, updated_at = NOW()
 WHERE id = 'max_yearly';
