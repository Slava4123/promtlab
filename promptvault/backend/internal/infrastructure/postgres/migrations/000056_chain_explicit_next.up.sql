-- Phase 16 v3 (Inline-tree editor): явные переходы между шагами.
--
-- До этой миграции переход для prompt-шага = неявно step с position+1 в той же
-- цепочке. Это работает для линейной цепочки, но при tree-структуре (fork с
-- независимыми ветками) приводит к «протечке» — последний шаг ветки A влетает
-- в первый шаг ветки B по позиции. Inline-tree editor требует явный граф.
--
-- Семантика после миграции:
--   prompt-шаг: переход = next_step_id (NULL = конец цепочки/ветки)
--   fork-шаг:   переход = выбранная юзером branch.next_step_id (как раньше)
--
-- Backfill: для каждого существующего prompt-шага ставим next_step_id = id шага
-- с position+1. Это сохраняет текущее поведение линейных цепочек 1:1. Для fork-
-- шагов next_step_id остаётся NULL — переход через conditions.branches.
--
-- ON DELETE SET NULL: при удалении шага все ссылающиеся next_step_id обнуляются
-- (концы цепочек обрываются). Service дополнительно re-link'ает предшественников
-- на T.next_step_id, чтобы цепочка «зашивалась».

ALTER TABLE prompt_chain_steps
    ADD COLUMN next_step_id BIGINT REFERENCES prompt_chain_steps(id) ON DELETE SET NULL;

-- Покрывает запросы вида «кто ссылается на этот шаг» при RemoveStep / re-link.
CREATE INDEX idx_prompt_chain_steps_next ON prompt_chain_steps (next_step_id)
    WHERE next_step_id IS NOT NULL;

UPDATE prompt_chain_steps p
   SET next_step_id = (
       SELECT n.id FROM prompt_chain_steps n
        WHERE n.chain_id = p.chain_id
          AND n.position = p.position + 1
        LIMIT 1
   )
 WHERE p.step_type = 'prompt';
