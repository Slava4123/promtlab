-- MN-78: explicit ON DELETE RESTRICT для prompt_chain_steps.prompt_id.
-- Раньше FK был без ON DELETE — PostgreSQL дефолтит на NO ACTION (alias
-- для RESTRICT), но это неявное поведение. Явное указание делает intent
-- очевидным: промпт нельзя удалить, если он используется в цепочке —
-- juzer должен сначала удалить шаг (или цепочку), либо chain.PurgeExpired
-- получит FK error 23503 (что и происходит, см. trash.PurgeExpired
-- который skip'ает promptIDs из prompt_chain_steps — pre-activation fix
-- из CLAUDE.md).
--
-- Поведенчески ничего не меняется (PG default = NO ACTION = RESTRICT
-- при immediate constraint), но grep'нув по миграциям юзер сразу видит
-- intent.

ALTER TABLE prompt_chain_steps
    DROP CONSTRAINT prompt_chain_steps_prompt_id_fkey;

ALTER TABLE prompt_chain_steps
    ADD CONSTRAINT prompt_chain_steps_prompt_id_fkey
    FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE RESTRICT;
