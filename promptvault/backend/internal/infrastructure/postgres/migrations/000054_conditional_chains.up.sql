-- Phase 16 Phase B: Conditional Chains. Расширение prompt_chain_steps для
-- условных ветвлений. Max-only tier check выполняется в Service.
--
-- step_type:
--   'prompt'      — обычный шаг (default, backward compat для существующих строк)
--   'conditional' — шаг-роутер: на основе output предыдущего шага выбирается ветка
--
-- conditions JSONB structure (только при step_type='conditional'):
-- {
--   "branches": [
--     {
--       "condition": {
--         "operator": "AND",  // "AND" | "OR" | "NOT" | "" (leaf with rules)
--         "rules": [
--           {"source": "step_<id>_output", "matcher": "contains", "value": "критический"}
--         ],
--         "children": []  // для вложенных AND/OR/NOT
--       },
--       "next_step_id": 42  // null для default branch (fallback если ничего не matched)
--     }
--   ]
-- }
--
-- Защита от ReDoS: Go regexp использует RE2 (linear time, no backtracking),
-- так что matcher="regex" не уязвим к catastrophic backtracking. Дополнительный
-- guard на длину pattern в evaluator (≤500 символов).

ALTER TABLE prompt_chain_steps
    ADD COLUMN step_type VARCHAR(20) NOT NULL DEFAULT 'prompt',
    ADD COLUMN conditions JSONB;

ALTER TABLE prompt_chain_steps
    ADD CONSTRAINT chk_prompt_chain_steps_type CHECK (step_type IN ('prompt', 'conditional'));
