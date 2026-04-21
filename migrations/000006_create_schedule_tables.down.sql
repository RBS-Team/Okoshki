DROP TRIGGER IF EXISTS update_schedule_exceptions_modtime ON master_schedule_exceptions;
DROP TABLE IF EXISTS master_schedule_exceptions;

DROP TRIGGER IF EXISTS update_working_hours_modtime ON master_working_hours;
DROP TABLE IF EXISTS master_working_hours;

ALTER TABLE master_services 
DROP COLUMN IF EXISTS is_auto_confirm;