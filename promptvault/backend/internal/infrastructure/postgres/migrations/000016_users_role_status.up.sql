-- Role / status fields for admin RBAC foundation.
-- Используем VARCHAR + CHECK вместо PG ENUM для совместимости с GORM AutoMigrate
-- (integration-тесты прогоняются на testcontainers через AutoMigrate). При этом
-- БД гарантирует целостность через CHECK constraint.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user',
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';

-- CHECK constraints добавляются отдельно (IF NOT EXISTS — PG 9.6+).
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_role_check'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('user', 'admin'));
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_status_check'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_status_check CHECK (status IN ('active', 'frozen'));
    END IF;
END $$;

-- Partial indexes: в системе будет 99% юзеров с role='user' / status='active',
-- индексы нужны только для быстрого поиска «необычных» записей.
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role) WHERE role != 'user';
CREATE INDEX IF NOT EXISTS idx_users_status ON users (status) WHERE status != 'active';
