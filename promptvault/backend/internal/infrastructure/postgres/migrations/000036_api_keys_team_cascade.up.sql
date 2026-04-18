-- M-36: hotfix для 000035 — меняем FK api_keys.team_id с SET NULL на CASCADE.
-- Идempotent: если 000035 в проде уже применилась с SET NULL, этот скрипт
-- заменит constraint. Если 000035 ещё не применялась (fresh install) — 000035
-- создаст FK уже с CASCADE, и этот скрипт выполнится как no-op с re-create.
-- Причина: при team DELETE SET NULL ограниченный ключ тихо становился unscoped
-- (privilege escalation, см. review MCP v1.1).
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_team_id_fkey;
ALTER TABLE api_keys
    ADD CONSTRAINT api_keys_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;
