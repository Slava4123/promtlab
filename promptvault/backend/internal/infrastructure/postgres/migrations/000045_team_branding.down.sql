ALTER TABLE teams
    DROP COLUMN IF EXISTS brand_primary_color,
    DROP COLUMN IF EXISTS brand_website,
    DROP COLUMN IF EXISTS brand_tagline,
    DROP COLUMN IF EXISTS brand_logo_url;
