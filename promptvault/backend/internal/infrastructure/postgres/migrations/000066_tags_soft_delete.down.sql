-- Rollback: дропаем индекс и колонку. Soft-deleted строки потеряются (станут
-- effectively visible) — это by design rollback'а.
DROP INDEX IF EXISTS idx_tags_deleted_at;
ALTER TABLE tags DROP COLUMN IF EXISTS deleted_at;
