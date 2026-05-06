-- Phase 16-Y Share TTL refactor: убираем active-count и daily-create квоты,
-- переходим на TTL-модель (как у Notion/Figma). Анти-абуз покрыт общим
-- per-user rate-limit на API (byUser(120/min) в routes.go).
--
-- Вектор миграции:
--   * share_links.expires_at: backfill legacy NULL → created_at+30d.
--     Колонка остаётся nullable: NULL = «бессрочно» (Max-эксклюзив).
--   * subscription_plans.max_share_links: больше не enforced — DROP.
--   * subscription_plans.max_daily_shares: больше не enforced — DROP.
--
-- Backward compat для frontend: поля исчезнут из /api/plans response;
-- pricing.tsx и share-dialog должны быть обновлены в той же раскатке.

UPDATE share_links
   SET expires_at = created_at + INTERVAL '30 days'
 WHERE expires_at IS NULL;

ALTER TABLE subscription_plans DROP COLUMN IF EXISTS max_share_links;
ALTER TABLE subscription_plans DROP COLUMN IF EXISTS max_daily_shares;
