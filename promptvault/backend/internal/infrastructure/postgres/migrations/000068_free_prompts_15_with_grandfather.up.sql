-- Pack E: понижаем Free лимит max_prompts с 50 → 15 (R8 из BUSINESS_RESEARCH:
-- 50 = "вечный жилец", конверсия Free→Pro 1.5%, цель 3-5%).
--
-- Grandfather-стратегия: НЕ ломаем существующих юзеров. Каждому текущему
-- Free-юзеру записываем snapshot старого лимита в users.legacy_quotas JSONB.
-- При проверке квоты Service.CheckPromptQuota использует legacy_quotas[field]
-- если есть, иначе plan.MaxPrompts. Новые регистрации получают пустой {} →
-- лимит из плана = 15.
--
-- Универсальный JSONB design: одна колонка покрывает любые будущие grandfather
-- сценарии (Pack F для max_ext_uses_daily/max_mcp_uses_daily использует ту же
-- инфраструктуру). Альтернатива «отдельная колонка на каждый лимит» — не
-- масштабируется и требует миграций при каждом изменении.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS legacy_quotas JSONB NOT NULL DEFAULT '{}'::jsonb;

-- Снапшотим текущий лимит Free плана для всех существующих Free-юзеров.
-- jsonb_build_object — типизированный builder, не требует ручного string-construction.
UPDATE users
   SET legacy_quotas = jsonb_build_object('max_prompts', 50)
 WHERE plan_id = 'free';

UPDATE subscription_plans
   SET max_prompts = 15, updated_at = NOW()
 WHERE id = 'free';
