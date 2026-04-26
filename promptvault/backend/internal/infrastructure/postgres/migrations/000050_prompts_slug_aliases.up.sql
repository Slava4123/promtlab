-- Phase 15: backward-compat для smysl-trahsliterirovannykh slug.
-- При изменении slug (например, ре-сluggified для cyrillic-titles) старый
-- сохраняется в slug_aliases jsonb-array. Share handler ищет в slug ИЛИ
-- slug_aliases @> '["<old>"]'::jsonb — старая опубликованная ссылка
-- продолжает резолвиться.
ALTER TABLE prompts
    ADD COLUMN IF NOT EXISTS slug_aliases jsonb NOT NULL DEFAULT '[]'::jsonb;

CREATE INDEX IF NOT EXISTS idx_prompts_slug_aliases_gin
    ON prompts USING GIN (slug_aliases)
    WHERE deleted_at IS NULL;
