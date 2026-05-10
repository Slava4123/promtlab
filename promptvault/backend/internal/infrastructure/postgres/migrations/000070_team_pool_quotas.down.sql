-- Откат: удаляем колонки team-pool квот. После применения rollback'а
-- логика квот возвращается к «всё считается per-user независимо от team_id»
-- (старое поведение до Pack T).

ALTER TABLE subscription_plans
    DROP COLUMN IF EXISTS max_team_prompts,
    DROP COLUMN IF EXISTS max_team_collections,
    DROP COLUMN IF EXISTS max_team_chains;
