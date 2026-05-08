-- MJ-40: переключаем uq_prompt_chain_steps_position с INITIALLY DEFERRED
-- на INITIALLY IMMEDIATE. Глобальная отложенность была overhead — все
-- INSERT/UPDATE проверки уникальности откладывались до COMMIT, даже
-- когда это не нужно (90% операций — single-row CREATE/DELETE step).
--
-- ReorderSteps репозиторий теперь явно ставит SET CONSTRAINTS DEFERRED
-- в начале transaction'а — локализует «отложенность» к нужному месту.

ALTER TABLE prompt_chain_steps
    ALTER CONSTRAINT uq_prompt_chain_steps_position INITIALLY IMMEDIATE;
