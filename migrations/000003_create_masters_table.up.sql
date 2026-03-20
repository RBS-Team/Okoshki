CREATE TABLE IF NOT EXISTS masters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    bio TEXT,
    avatar_url VARCHAR(255),
    timezone VARCHAR(50) DEFAULT 'Europe/Moscow',
    lat DECIMAL(10, 7),
    lon DECIMAL(10, 7),
    rating DECIMAL(3, 2) DEFAULT 0,
    review_count INT DEFAULT 0,
    reports_count INT DEFAULT 0,
    is_blocked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_masters_user_id ON masters(user_id);
CREATE INDEX IF NOT EXISTS idx_masters_rating ON masters(rating DESC);

CREATE TRIGGER update_masters_modtime
BEFORE UPDATE ON masters 
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();