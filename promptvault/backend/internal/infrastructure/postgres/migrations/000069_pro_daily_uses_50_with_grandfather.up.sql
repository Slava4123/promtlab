-- Pack F: понижаем Pro дневные лимиты max_ext_uses_daily и max_mcp_uses_daily
-- с 100 → 50. Обоснование: 100/день = по одной вставке каждые 5 минут весь
-- 8-часовой день — soft-безлимит, нет конверсионного давления к Max. 50 даёт
-- комфортный day-job, но активный power-user через 1-2 месяца упирается →
-- strong upsell к Max (500/день).
--
-- Grandfather-стратегия: каждому текущему Pro-юзеру (включая pro_yearly)
-- записываем snapshot старых лимитов в users.legacy_quotas (использует ту же
-- инфраструктуру что Pack E, миграция 000068).
--
-- jsonb_set с merge: НЕ затираем существующие ключи (если у юзера уже есть
-- legacy_quotas{"max_prompts":50} от Pack E, мы добавим max_ext_uses_daily
-- сверху, не удалив max_prompts). Используем оператор ||: правый операнд
-- перезаписывает левый только по совпадающим ключам, новые добавляются.

UPDATE users
   SET legacy_quotas = legacy_quotas || jsonb_build_object(
       'max_ext_uses_daily', 100,
       'max_mcp_uses_daily', 100
   )
 WHERE plan_id IN ('pro', 'pro_yearly');

UPDATE subscription_plans
   SET max_ext_uses_daily = 50,
       max_mcp_uses_daily = 50,
       updated_at = NOW()
 WHERE id IN ('pro', 'pro_yearly');
