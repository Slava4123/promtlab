-- ВНИМАНИЕ: вернуть NOT NULL можно, только если все fork-шаги имеют prompt_id.
-- В норме после v3.1 fork-шаги имеют NULL — этот rollback заблокируется БД,
-- если существуют fork-шаги без промпта. Это намеренно: rollback не должен
-- молча терять данные.

ALTER TABLE prompt_chain_steps
    ALTER COLUMN prompt_id SET NOT NULL;
