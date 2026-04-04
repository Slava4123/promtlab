-- 000001_initial_schema.down.sql
-- Drop all tables in reverse dependency order.

DROP TABLE IF EXISTS prompt_collections;
DROP TABLE IF EXISTS prompt_tags;
DROP TABLE IF EXISTS prompt_versions;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS prompts;
DROP TABLE IF EXISTS collections;
DROP TABLE IF EXISTS team_invitations;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS email_verifications;
DROP TABLE IF EXISTS linked_accounts;
DROP TABLE IF EXISTS users;
