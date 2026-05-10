-- Phase 16 v3.1 backfill: fork-шаг — контейнер без своего промпта (см. 000057).
-- Если в БД уже есть fork-шаги, созданные до v3.1 (с непустым prompt_id) —
-- зануляем prompt_id, чтобы новый UI не показывал их как «обычный шаг с промптом».
-- На dark-launch (CHAINS_ENABLED=false) prod-таблица обычно пуста; это safety net.

UPDATE prompt_chain_steps
   SET prompt_id = NULL
 WHERE step_type = 'fork'
   AND prompt_id IS NOT NULL;
