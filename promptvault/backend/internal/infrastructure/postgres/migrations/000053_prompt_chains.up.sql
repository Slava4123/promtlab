-- Phase 16: Prompt Chains — связанные последовательности промптов с output→input маппингом.
-- Conditional Chains (step_type, conditions) появятся в миграции 000054 (Phase B).
--
-- Таблицы:
--   prompt_chains            — мета-данные цепочки (имя, владелец, soft-delete)
--   prompt_chain_steps       — упорядоченные шаги (ссылка на prompt + variable_mapping)
--   prompt_chain_executions  — запуски цепочки (current_step, snapshot структуры)
--
-- prompt_id без ON DELETE CASCADE: удаление промпта, используемого в цепочке,
-- блокируется на уровне сервиса (UseCase.DeletePrompt → 409 Conflict).
-- BUSINESS_RESEARCH рекомендует unique numbers вместо -1 sentinel (миграция 000046).

CREATE TABLE prompt_chains (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id     BIGINT REFERENCES teams(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_prompt_chains_user_id ON prompt_chains (user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_prompt_chains_team_id ON prompt_chains (team_id) WHERE deleted_at IS NULL AND team_id IS NOT NULL;

CREATE TABLE prompt_chain_steps (
    id                BIGSERIAL PRIMARY KEY,
    chain_id          BIGINT NOT NULL REFERENCES prompt_chains(id) ON DELETE CASCADE,
    position          INTEGER NOT NULL,
    prompt_id         BIGINT NOT NULL REFERENCES prompts(id),
    name              VARCHAR(255) NOT NULL DEFAULT '',
    variable_mapping  JSONB NOT NULL DEFAULT '{}'::jsonb,
    manual_checkpoint BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- DEFERRABLE: позволяет ReorderSteps временно нарушать уникальность внутри транзакции
-- (UPDATE position для нескольких строк подряд) — проверка отложена до COMMIT.
ALTER TABLE prompt_chain_steps
    ADD CONSTRAINT uq_prompt_chain_steps_position UNIQUE (chain_id, position) DEFERRABLE INITIALLY DEFERRED;

CREATE INDEX idx_prompt_chain_steps_prompt ON prompt_chain_steps (prompt_id);

CREATE TABLE prompt_chain_executions (
    id             BIGSERIAL PRIMARY KEY,
    chain_id       BIGINT NOT NULL REFERENCES prompt_chains(id) ON DELETE CASCADE,
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_step   INTEGER NOT NULL DEFAULT 1,
    variables      JSONB NOT NULL DEFAULT '{}'::jsonb,
    step_outputs   JSONB NOT NULL DEFAULT '{}'::jsonb,
    chain_snapshot JSONB NOT NULL,
    status         VARCHAR(20) NOT NULL DEFAULT 'in_progress'
                   CHECK (status IN ('in_progress', 'completed', 'abandoned')),
    started_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at   TIMESTAMPTZ,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prompt_chain_executions_user ON prompt_chain_executions (user_id, status);
CREATE INDEX idx_prompt_chain_executions_chain ON prompt_chain_executions (chain_id);
-- Алерт ChainExecutionStuck: status='in_progress' AND updated_at < now() - 24h.
CREATE INDEX idx_prompt_chain_executions_active ON prompt_chain_executions (updated_at) WHERE status = 'in_progress';

-- Tier-лимиты для chains (паттерн 000046: конкретные числа, не sentinel).
-- Числа подобраны по аналогии max_prompts/max_collections, чтобы Max был enough для 99% юзеров.
ALTER TABLE subscription_plans
    ADD COLUMN max_chains           INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN max_steps_per_chain  INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN max_saved_executions INTEGER NOT NULL DEFAULT 0;

UPDATE subscription_plans
   SET max_chains = 1,
       max_steps_per_chain = 3,
       max_saved_executions = 0,
       updated_at = NOW()
 WHERE id = 'free';

UPDATE subscription_plans
   SET max_chains = 5,
       max_steps_per_chain = 10,
       max_saved_executions = 10,
       updated_at = NOW()
 WHERE id IN ('pro', 'pro_yearly');

UPDATE subscription_plans
   SET max_chains = 100,
       max_steps_per_chain = 50,
       max_saved_executions = 1000,
       updated_at = NOW()
 WHERE id IN ('max', 'max_yearly');
