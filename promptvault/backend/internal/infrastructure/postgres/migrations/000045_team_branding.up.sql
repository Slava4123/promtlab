-- Phase 14 Branded share pages (Max-only). Публичные страницы /s/:token
-- для команд на Max могут показать кастомный brand: logo, tagline, website,
-- primary color. Все поля nullable — существующие записи не затронуты.
--
-- Ограничения валидации (на стороне application, не DB):
--   - logo_url / website: HTTPS-only, max 500 символов
--   - tagline: plain text (XSS защитит React эскейпингом), max 200
--   - primary_color: #RRGGBB формат (hex)

ALTER TABLE teams
    ADD COLUMN IF NOT EXISTS brand_logo_url       VARCHAR(500),
    ADD COLUMN IF NOT EXISTS brand_tagline        VARCHAR(200),
    ADD COLUMN IF NOT EXISTS brand_website        VARCHAR(500),
    ADD COLUMN IF NOT EXISTS brand_primary_color  VARCHAR(7);
