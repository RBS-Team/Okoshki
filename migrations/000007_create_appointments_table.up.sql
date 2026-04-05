CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS appointments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL REFERENCES "user"(user_id) ON DELETE CASCADE,
    master_id UUID NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES master_services(id) ON DELETE RESTRICT,
    start_at TIMESTAMP WITH TIME ZONE NOT NULL,
    end_at TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    is_manual_block BOOLEAN DEFAULT false,
    client_comment TEXT,
    master_note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT chk_appointments_time_order CHECK (start_at < end_at),
    
    CONSTRAINT chk_appointments_status CHECK (
        status IN ('pending', 'confirmed', 'rejected', 'cancelled', 'completed')
    ),

    CONSTRAINT exclude_overlapping_appointments EXCLUDE USING gist (
        master_id WITH =,
        tstzrange(start_at, end_at) WITH &&
    ) WHERE (status IN ('pending', 'confirmed'))
);

CREATE INDEX IF NOT EXISTS idx_appointments_master_id_start_at ON appointments(master_id, start_at);
CREATE INDEX IF NOT EXISTS idx_appointments_client_id ON appointments(client_id);

CREATE TRIGGER update_appointments_modtime
BEFORE UPDATE ON appointments 
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();