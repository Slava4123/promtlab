-- Prefix-optimized indexes for autocomplete suggestions.
-- text_pattern_ops allows B-tree to satisfy LIKE 'prefix%' queries on lowered text.
CREATE INDEX IF NOT EXISTS idx_prompts_title_prefix
  ON prompts (user_id, lower(title) text_pattern_ops)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_collections_name_prefix
  ON collections (user_id, lower(name) text_pattern_ops)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_tags_name_prefix
  ON tags (user_id, lower(name) text_pattern_ops);
