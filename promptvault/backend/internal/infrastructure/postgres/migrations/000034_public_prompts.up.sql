-- Public prompts: чекбокс "публичный" + SEO-friendly URL /p/:slug.
-- Отличается от share-link (/s/:token):
--   share-link — приватная ссылка по токену, для команды/друзей;
--   public     — SEO-индексируемое представление, для привлечения трафика.
-- is_public default false, slug NULL пока не сделан публичным.
ALTER TABLE prompts
    ADD COLUMN IF NOT EXISTS is_public BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS slug      VARCHAR(120);

-- UNIQUE на slug там где он проставлен (public только). NULL-slug'и не мешают.
CREATE UNIQUE INDEX IF NOT EXISTS idx_prompts_slug
    ON prompts (slug) WHERE slug IS NOT NULL;

-- Индекс для sitemap/list публичных — фильтр по is_public=true + deleted_at IS NULL.
CREATE INDEX IF NOT EXISTS idx_prompts_public
    ON prompts (updated_at DESC) WHERE is_public = TRUE AND deleted_at IS NULL;
