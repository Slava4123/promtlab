-- Phase 16-X Branding UX: загрузка логотипа файлом (бэк-storage = bytea).
-- Дополняет 000045_team_branding: вместо/в дополнение к URL-полю owner Max-команды
-- может загрузить PNG/JPEG/WebP файл, который хранится в bytea и отдаётся через
-- GET /api/teams/{slug}/branding/logo с ETag (sha256) + 24h immutable cache.
--
-- Backward compat: brand_logo_source default 'url' для всех существующих команд;
-- если у команды brand_logo_url пустой — фронт покажет «нет логотипа», источник
-- остаётся 'url'. После загрузки файла источник переключается на 'file' атомарно
-- в одном usecase-вызове с upsert'ом team_logo_files.
--
-- Лимиты application-level (см. usecases/team/logo.go):
--   - размер: ≤1 МБ (1 048 576 байт)
--   - формат: image/png, image/jpeg, image/webp (whitelist, без SVG = XSS)
--   - размеры: width/height ≤1024px

ALTER TABLE teams
    ADD COLUMN IF NOT EXISTS brand_logo_source TEXT NOT NULL DEFAULT 'url'
    CHECK (brand_logo_source IN ('url', 'file', 'none'));

CREATE TABLE IF NOT EXISTS team_logo_files (
    team_id      BIGINT      PRIMARY KEY REFERENCES teams(id) ON DELETE CASCADE,
    content_type TEXT        NOT NULL CHECK (content_type IN ('image/png','image/jpeg','image/webp')),
    size_bytes   BIGINT      NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 1048576),
    sha256       CHAR(64)    NOT NULL,
    bytes        BYTEA       NOT NULL,
    uploaded_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_team_logo_files_sha256 ON team_logo_files(sha256);
