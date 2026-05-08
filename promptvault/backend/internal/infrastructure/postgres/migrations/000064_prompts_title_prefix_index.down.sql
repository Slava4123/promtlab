-- migrate:no-transaction
DROP INDEX CONCURRENTLY IF EXISTS idx_prompts_user_title_lower_prefix;
DROP INDEX CONCURRENTLY IF EXISTS idx_prompts_team_title_lower_prefix;
