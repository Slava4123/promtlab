ALTER TABLE collections
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_collections_deleted_at ON collections (deleted_at);
