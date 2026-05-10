-- Откат rename: 'fork' → 'conditional'. Безопасен только если caller вернёт
-- старый код с DSL evaluator. label-поля в JSONB не удаляются (их игнорирует
-- старый код при unmarshal в Condition struct).

ALTER TABLE prompt_chain_steps
    DROP CONSTRAINT IF EXISTS chk_prompt_chain_steps_type;

UPDATE prompt_chain_steps SET step_type = 'conditional' WHERE step_type = 'fork';

ALTER TABLE prompt_chain_steps
    ADD CONSTRAINT chk_prompt_chain_steps_type CHECK (step_type IN ('prompt', 'conditional'));
