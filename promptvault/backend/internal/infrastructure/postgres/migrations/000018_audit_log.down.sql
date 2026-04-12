DROP TRIGGER IF EXISTS audit_log_prevent_delete ON audit_log;
DROP TRIGGER IF EXISTS audit_log_prevent_update ON audit_log;
DROP FUNCTION IF EXISTS prevent_audit_log_modification();
DROP TABLE IF EXISTS audit_log;
