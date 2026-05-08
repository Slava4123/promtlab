ALTER TABLE prompt_chain_steps
    DROP CONSTRAINT prompt_chain_steps_prompt_id_fkey;

ALTER TABLE prompt_chain_steps
    ADD CONSTRAINT prompt_chain_steps_prompt_id_fkey
    FOREIGN KEY (prompt_id) REFERENCES prompts(id);
