-- Phase 16 v2 (Tree-canvas): rename step_type 'conditional' → 'fork'.
-- DSL evaluator (matchers / AND/OR/NOT) удалён, теперь юзер сам выбирает ветку
-- по её Label в run-mode UI.
--
-- Структура conditions JSONB меняется:
--   старая: {branches: [{condition: {rules: [...]}, next_step_id: N}]}
--   новая:  {branches: [{label: "Если ...", next_step_id: N}]}
-- Если в проде уже есть conditional-шаги — у них нет label'ов, после миграции
-- придётся вручную переименовать ветки через Canvas. На dark launch
-- CHAINS_ENABLED=false БД пустая по этому полю — миграция данных не нужна.

ALTER TABLE prompt_chain_steps
    DROP CONSTRAINT IF EXISTS chk_prompt_chain_steps_type;

UPDATE prompt_chain_steps SET step_type = 'fork' WHERE step_type = 'conditional';

ALTER TABLE prompt_chain_steps
    ADD CONSTRAINT chk_prompt_chain_steps_type CHECK (step_type IN ('prompt', 'fork'));
