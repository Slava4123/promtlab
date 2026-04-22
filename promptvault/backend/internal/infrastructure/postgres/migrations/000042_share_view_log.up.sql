-- share_view_log — timeline просмотров публичных шар-ссылок для Pro+ аналитики
-- (Phase 14, задача 2). Даёт разбивку по дням, рефереру и стране.
--
-- Запись ТОЛЬКО для владельцев на тарифе Pro/Max (проверка в
-- delivery/http/share/public.go: если owner.plan_id == 'free' — skip INSERT).
-- Free: в UI видят только total view_count из share_links.view_count
-- (существующее поле, апсейл на Pro даст timeline).
--
-- country CHAR(2) — ISO-3166-1 alpha-2 (опционально, заполняется из GeoIP
-- в Phase D. Сейчас всегда NULL, поле готово).
--
-- Retention: cleanup cron (Phase A.8) удаляет записи по max retention плана
-- владельца share-ссылки (Pro=90 дней, Max=365 дней). DELETE разрешён.

CREATE TABLE IF NOT EXISTS share_view_log (
    id                BIGSERIAL   PRIMARY KEY,
    share_link_id     BIGINT      NOT NULL REFERENCES share_links(id) ON DELETE CASCADE,
    viewed_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    referer           VARCHAR(500),
    country           CHAR(2),
    user_agent_family VARCHAR(50)
);

-- Основной query: timeline просмотров конкретной ссылки.
CREATE INDEX IF NOT EXISTS idx_svl_link_viewed
    ON share_view_log (share_link_id, viewed_at DESC);

-- Для cleanup cron: WHERE viewed_at < NOW() - INTERVAL '...'
CREATE INDEX IF NOT EXISTS idx_svl_viewed
    ON share_view_log (viewed_at);
