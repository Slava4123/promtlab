-- Откат 000054. Безопасен только если ни один шаг не имеет step_type='conditional'.
-- Если есть — conditional конфигурации теряются, шаги "деградируют" в обычные prompt.
-- Caller должен проверить заранее: SELECT count(*) FROM prompt_chain_steps WHERE step_type = 'conditional';

ALTER TABLE prompt_chain_steps
    DROP CONSTRAINT IF EXISTS chk_prompt_chain_steps_type;

ALTER TABLE prompt_chain_steps
    DROP COLUMN IF EXISTS conditions,
    DROP COLUMN IF EXISTS step_type;
