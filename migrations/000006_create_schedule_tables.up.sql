ALTER TABLE master_services 
ADD COLUMN IF NOT EXISTS is_auto_confirm BOOLEAN DEFAULT true;

CREATE TABLE IF NOT EXISTS master_working_hours (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    master_id UUID NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    start_time TIME,
    end_time TIME,
    is_day_off BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_working_time_order CHECK (
        (is_day_off = true AND start_time IS NULL AND end_time IS NULL) OR 
        (is_day_off = false AND start_time IS NOT NULL AND end_time IS NOT NULL AND start_time < end_time)
    ),
    UNIQUE(master_id, day_of_week)
);

CREATE INDEX IF NOT EXISTS idx_working_hours_master_id ON master_working_hours(master_id);

CREATE TRIGGER update_working_hours_modtime
BEFORE UPDATE ON master_working_hours 
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

CREATE TABLE IF NOT EXISTS master_schedule_exceptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    master_id UUID NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    exception_date DATE NOT NULL,
    start_time TIME,
    end_time TIME,
    is_working BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_exception_time_order CHECK (
        (is_working = false AND start_time IS NULL AND end_time IS NULL) OR 
        (is_working = true AND start_time IS NOT NULL AND end_time IS NOT NULL AND start_time < end_time)
    ),
    UNIQUE(master_id, exception_date)
);

CREATE INDEX IF NOT EXISTS idx_schedule_exceptions_master_date ON master_schedule_exceptions(master_id, exception_date);

CREATE TRIGGER update_schedule_exceptions_modtime
BEFORE UPDATE ON master_schedule_exceptions 
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();