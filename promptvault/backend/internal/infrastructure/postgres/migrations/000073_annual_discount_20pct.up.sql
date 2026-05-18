-- Annual discount: 10% → 20%.
-- Существующие подписки НЕ затрагиваются (T-Bank rebillId identifies card, not amount).
-- На renewal `plans.GetByID()` прочтёт новую цену → T-Bank Charge на новую сумму.
--   pro_yearly:  6490 → 5750 ₽ (-20% от monthly×12 = 7188)
--   max_yearly: 13990 → 12470 ₽ (-20% от monthly×12 = 15588)
UPDATE subscription_plans
   SET price_kop = 575000, updated_at = NOW()
 WHERE id = 'pro_yearly';

UPDATE subscription_plans
   SET price_kop = 1247000, updated_at = NOW()
 WHERE id = 'max_yearly';
