-- Phase 15 polish: статус отзыва для admin-страницы /admin/feedbacks.
-- new (default) → отзыв ещё не прочитан админом.
-- read           → админ открыл detail и пометил как прочитанный.
-- archived       → админ убрал из основного списка (не удалено, можно вернуть).
CREATE TYPE feedback_status AS ENUM ('new', 'read', 'archived');

ALTER TABLE feedbacks
    ADD COLUMN status feedback_status NOT NULL DEFAULT 'new';

-- Индекс для частого фильтра list-страницы по статусу.
CREATE INDEX IF NOT EXISTS idx_feedbacks_status ON feedbacks (status);
