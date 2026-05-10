DROP INDEX IF EXISTS idx_prompt_chain_steps_next;

ALTER TABLE prompt_chain_steps
    DROP COLUMN IF EXISTS next_step_id;
