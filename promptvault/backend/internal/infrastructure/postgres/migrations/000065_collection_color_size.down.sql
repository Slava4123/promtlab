-- Rollback: вернуть VARCHAR(20). Backfill не отменяется (значения, заменённые
-- на default #8b5cf6, остаются — оригинальные значения утеряны).
ALTER TABLE collections
ALTER COLUMN color TYPE VARCHAR(20);
