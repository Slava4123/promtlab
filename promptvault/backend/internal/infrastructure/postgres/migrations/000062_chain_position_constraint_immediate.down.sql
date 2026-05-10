ALTER TABLE prompt_chain_steps
    DROP CONSTRAINT IF EXISTS uq_prompt_chain_steps_position;

ALTER TABLE prompt_chain_steps
    ADD CONSTRAINT uq_prompt_chain_steps_position UNIQUE (chain_id, position) DEFERRABLE INITIALLY DEFERRED;
