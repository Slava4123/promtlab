-- Phase 16-Y: Add description column to prompts.
-- Background:
--   Extension UI editor + form schema + API request body уже шлёт `description`,
--   но backend INSERT/UPDATE никогда не записывал — колонки в БД не было.
--   Это давало silent data loss: юзер вводит описание промпта, на refresh оно
--   исчезает. Найдено в self-review 2026-05-16 (см. B8).
-- Why varchar(2000):
--   Согласуется с CreatePromptBody validation limit (max=2000 в shared zod).
--   Default '' — соответствует поведению collection.description (NOT NULL "").

ALTER TABLE prompts
    ADD COLUMN IF NOT EXISTS description varchar(2000) NOT NULL DEFAULT '';
