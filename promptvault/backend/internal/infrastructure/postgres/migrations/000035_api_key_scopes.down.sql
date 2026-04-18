DROP INDEX IF EXISTS idx_api_keys_expires_at;
ALTER TABLE api_keys
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS allowed_tools,
    DROP COLUMN IF EXISTS team_id,
    DROP COLUMN IF EXISTS read_only;
