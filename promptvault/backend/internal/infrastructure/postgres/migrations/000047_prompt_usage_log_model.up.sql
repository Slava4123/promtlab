-- Phase 14.2 — segmentation по модели AI: новая колонка prompt_usage_log.model_used.
-- При каждом IncrementUsage запоминаем модель промпта (Claude/GPT/DeepSeek/…)
-- чтобы /analytics мог строить pie-chart "по каким моделям юзер работает".
--
-- Старые записи получают NULL (будут показываться как "Без модели" в UI).

ALTER TABLE prompt_usage_log
    ADD COLUMN IF NOT EXISTS model_used VARCHAR(50);

CREATE INDEX IF NOT EXISTS idx_pul_model_used
    ON prompt_usage_log(model_used)
    WHERE model_used IS NOT NULL;
