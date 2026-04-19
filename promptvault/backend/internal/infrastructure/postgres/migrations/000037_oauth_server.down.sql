-- Rollback OAuth 2.1 Authorization Server.
-- Порядок важен: tokens и codes ссылаются на clients через FK.
DROP TABLE IF EXISTS oauth_tokens;
DROP TABLE IF EXISTS oauth_authorization_codes;
DROP TABLE IF EXISTS oauth_clients;
