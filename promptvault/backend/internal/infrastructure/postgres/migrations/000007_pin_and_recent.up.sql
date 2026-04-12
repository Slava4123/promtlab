-- Таблица пинов промптов (личные + командные)
CREATE TABLE IF NOT EXISTS prompt_pins (
    id           BIGSERIAL PRIMARY KEY,
    prompt_id    BIGINT NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_team_wide BOOLEAN NOT NULL DEFAULT FALSE,
    pinned_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Личный пин: один на пользователя на промпт
CREATE UNIQUE INDEX idx_prompt_pins_personal
    ON prompt_pins (prompt_id, user_id) WHERE is_team_wide = FALSE;

-- Командный пин: один на промпт (кто бы ни поставил)
CREATE UNIQUE INDEX idx_prompt_pins_team
    ON prompt_pins (prompt_id) WHERE is_team_wide = TRUE;

-- Для быстрого запроса "все пины юзера"
CREATE INDEX idx_prompt_pins_user_id ON prompt_pins (user_id);

-- Колонка last_used_at для секции "Недавние"
ALTER TABLE prompts ADD COLUMN last_used_at TIMESTAMPTZ;

-- Индекс для сортировки по last_used_at (только живые промпты)
CREATE INDEX idx_prompts_last_used_at
    ON prompts (last_used_at DESC NULLS LAST) WHERE deleted_at IS NULL;
