-- Откат миграции 000053. Безопасен только если ещё нет пользовательских данных в chains.
-- Если есть данные — снапшоты из prompt_chain_executions и созданные цепочки потеряются.

ALTER TABLE subscription_plans
    DROP COLUMN IF EXISTS max_saved_executions,
    DROP COLUMN IF EXISTS max_steps_per_chain,
    DROP COLUMN IF EXISTS max_chains;

DROP TABLE IF EXISTS prompt_chain_executions;
DROP TABLE IF EXISTS prompt_chain_steps;
DROP TABLE IF EXISTS prompt_chains;
