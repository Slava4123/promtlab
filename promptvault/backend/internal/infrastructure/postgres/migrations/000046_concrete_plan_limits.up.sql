-- Concrete plan limits: заменяем все -1 ("безлимит") на реальные значения.
-- Продуктовое решение: на странице /pricing и в UI показываем конкретные цифры,
-- а не слово "безлимит" — юзер видит что ценность тарифа измерима. Числа подобраны
-- так, чтобы Max был enough для 99% пользователей, но при этом видимый потолок есть.
--
-- Таблица значений:
--   free: 50 promptов, 3 коллекции, 1 команда (до 3 участников),
--         2 share-ссылки total, 10 share-ссылок/день, 5 ext/day, 5 mcp/day
--   pro:  500 / 100 / 5 (до 10 участников) / 50 / 100 / 100 / 100
--   max:  10000 / 1000 / 50 (до 50 участников) / 500 / 1000 / 500 / 500
--
-- down-миграция возвращает -1 только для тех полей где они были (согласно 000019).

UPDATE subscription_plans SET
    max_prompts      = 500,
    max_collections  = 100,
    max_teams        = 5,
    max_team_members = 10,
    max_share_links  = 50
WHERE id IN ('pro', 'pro_yearly');

UPDATE subscription_plans SET
    max_prompts       = 10000,
    max_collections   = 1000,
    max_teams         = 50,
    max_team_members  = 50,
    max_share_links   = 500,
    max_ext_uses_daily = 500,
    max_mcp_uses_daily = 500
WHERE id IN ('max', 'max_yearly');

-- Free-плана значения не трогаем (они уже конкретные, не -1).
