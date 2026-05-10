-- MN-34: parity с collections.deleted_at. Tag soft-delete позволяет восстановить
-- случайно удалённый тег вместе с привязкой к промптам (через prompt_tags пары
-- остаются — restore через unscoped query).
ALTER TABLE tags ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Partial index только на soft-deleted — основная масса запросов идёт через
-- WHERE deleted_at IS NULL (gorm.DeletedAt scope), не нужен полный b-tree.
CREATE INDEX IF NOT EXISTS idx_tags_deleted_at ON tags (deleted_at) WHERE deleted_at IS NOT NULL;
