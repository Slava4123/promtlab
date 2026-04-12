DROP INDEX IF EXISTS idx_collections_deleted_at;

ALTER TABLE collections
    DROP COLUMN IF EXISTS deleted_at;
