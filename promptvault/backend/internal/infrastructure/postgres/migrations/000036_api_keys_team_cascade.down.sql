-- Откат возвращает SET NULL (историческое поведение 000035).
-- Важно: rollback НЕ рекомендуется для prod — восстанавливает уязвимость.
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_team_id_fkey;
ALTER TABLE api_keys
    ADD CONSTRAINT api_keys_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL;
