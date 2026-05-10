-- Pack T (Team-pool квоты): отделяем командные ресурсы от персональных.
-- До этой миграции CountPrompts(userID) возвращал ВСЕ промпты юзера —
-- включая те что он создал в командах. Free участник Pro команды упирался
-- в свой personal лимит 15 промптов и блокировал команду.
--
-- После миграции:
--   personal квота — только prompts с team_id IS NULL (соло-юзера)
--   team квота    — pool на всю команду, лимит из плана owner'а команды
--
-- Числа подобраны по принципу «команда — это buff, не downgrade»:
-- Pro team_prompts = 4× personal (2000 vs 500). Это даёт реальную team-value
-- но не превращает Pro в безлимитный workspace (для агентств 10+ — Team tier 2999₽).
-- Защита от abuse через max_team_members уже встроена (Pro=10, Max=50).

ALTER TABLE subscription_plans
    ADD COLUMN max_team_prompts     INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN max_team_collections INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN max_team_chains      INTEGER NOT NULL DEFAULT 0;

UPDATE subscription_plans SET
    max_team_prompts     = 50,
    max_team_collections = 10,
    max_team_chains      = 3,
    updated_at = NOW()
 WHERE id = 'free';

UPDATE subscription_plans SET
    max_team_prompts     = 2000,
    max_team_collections = 400,
    max_team_chains      = 20,
    updated_at = NOW()
 WHERE id IN ('pro', 'pro_yearly');

UPDATE subscription_plans SET
    max_team_prompts     = 50000,
    max_team_collections = 5000,
    max_team_chains      = 500,
    updated_at = NOW()
 WHERE id IN ('max', 'max_yearly');
