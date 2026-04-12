DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_role;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_status_check,
    DROP CONSTRAINT IF EXISTS users_role_check,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS role;
