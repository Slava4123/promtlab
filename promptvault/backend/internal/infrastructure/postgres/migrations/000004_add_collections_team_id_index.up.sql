CREATE INDEX IF NOT EXISTS idx_collections_team_id ON collections (team_id) WHERE team_id IS NOT NULL;
