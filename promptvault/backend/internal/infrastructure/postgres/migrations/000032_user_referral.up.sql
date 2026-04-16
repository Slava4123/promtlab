-- M-7: реферальная программа.
-- referral_code — уникальный 8-символьный код юзера (base32 + проверка на слова).
--   Генерируется при создании аккаунта (auth.Register, oauth create).
-- referred_by — код пригласившего (nullable). Записывается из ?ref=XXXXX cookie
--   на момент регистрации/первого OAuth-callback.
-- referral_rewarded_at — момент выдачи награды рефереру. NOT NULL после первой
--   успешной оплаты этого юзера → используется для idempotency, чтобы не
--   наградить рефера несколько раз при повторных платежах того же рефери.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS referral_code        VARCHAR(16),
    ADD COLUMN IF NOT EXISTS referred_by          VARCHAR(16),
    ADD COLUMN IF NOT EXISTS referral_rewarded_at TIMESTAMPTZ;

-- Заполняем существующих юзеров случайными кодами, чтобы UNIQUE constraint
-- не упал. Используем substr(md5(...), ...) — достаточно случайно для backfill,
-- новые юзеры получают более энтропийные коды через Go-генератор.
UPDATE users
   SET referral_code = UPPER(SUBSTRING(MD5(id::text || 'pv_ref_seed' || extract(epoch from created_at)::text) FOR 8))
 WHERE referral_code IS NULL;

-- После backfill ставим NOT NULL + UNIQUE.
ALTER TABLE users
    ALTER COLUMN referral_code SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_referral_code
    ON users (referral_code);

-- Индекс для поиска рефералов конкретного юзера ("кто пригласил X": GET /referral).
CREATE INDEX IF NOT EXISTS idx_users_referred_by
    ON users (referred_by) WHERE referred_by IS NOT NULL;
