DROP TRIGGER IF EXISTS team_activity_log_prevent_mutation ON team_activity_log;
DROP FUNCTION IF EXISTS prevent_team_activity_log_mutation();
DROP TABLE IF EXISTS team_activity_log;
