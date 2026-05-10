-- MJ-40: переключаем uq_prompt_chain_steps_position с INITIALLY DEFERRED
-- на INITIALLY IMMEDIATE. Глобальная отложенность была overhead — все
-- INSERT/UPDATE проверки уникальности откладывались до COMMIT, даже
-- когда это не нужно (90% операций — single-row CREATE/DELETE step).
--
-- ReorderSteps репозиторий теперь явно ставит SET CONSTRAINTS DEFERRED
-- в начале transaction'а — локализует «отложенность» к нужному месту.
--
-- DROP + ADD: PostgreSQL поддерживает `ALTER CONSTRAINT ... INITIALLY ...`
-- только для FOREIGN KEY constraints. Для UNIQUE constraint — drop + recreate.
-- Constraint остаётся DEFERRABLE, чтобы ReorderSteps мог SET CONSTRAINTS DEFERRED.

ALTER TABLE prompt_chain_steps
    DROP CONSTRAINT IF EXISTS uq_prompt_chain_steps_position;

ALTER TABLE prompt_chain_steps
    ADD CONSTRAINT uq_prompt_chain_steps_position UNIQUE (chain_id, position) DEFERRABLE INITIALLY IMMEDIATE;
