CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- users

CREATE TABLE IF NOT EXISTS "user" (
    user_id       UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(50)  NOT NULL DEFAULT 'client',
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_email ON "user"(email);

-- categories

CREATE TABLE IF NOT EXISTS category (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    avatar_url VARCHAR(255),
    is_active   BOOLEAN      DEFAULT TRUE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_category_modtime
BEFORE UPDATE ON category
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- masters

CREATE TABLE IF NOT EXISTS masters (
    id            UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID          NOT NULL UNIQUE,
    category_id   UUID          NOT NULL REFERENCES category(id),
    first_name    VARCHAR(255)  NOT NULL,
    last_name     VARCHAR(255)  NOT NULL,
    phone         VARCHAR(20)   NOT NULL CHECK (phone ~ '^\+[1-9][0-9]{6,14}$'),
    address       VARCHAR(255)  NOT NULL,
    city          VARCHAR(255)  NOT NULL,
    bio           TEXT,
    avatar_url    VARCHAR(255),
    timezone      VARCHAR(50)   DEFAULT 'Europe/Moscow',
    lat           DOUBLE PRECISION,
    lon           DOUBLE PRECISION,
    rating        DECIMAL(3, 2) DEFAULT 0,
    review_count  INT           DEFAULT 0,
    reports_count INT           DEFAULT 0,
    is_blocked    BOOLEAN       DEFAULT FALSE,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_masters_rating ON masters(rating DESC);
CREATE INDEX IF NOT EXISTS idx_masters_category_rating ON masters(category_id, rating DESC);

CREATE TRIGGER update_masters_modtime
BEFORE UPDATE ON masters
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- master_services

CREATE TABLE IF NOT EXISTS master_services (
    id                    UUID           PRIMARY KEY DEFAULT uuid_generate_v4(),
    master_id             UUID           NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    category_id           UUID           NOT NULL REFERENCES category(id) ON DELETE RESTRICT,
    title                 VARCHAR(255)   NOT NULL,
    address               TEXT           NOT NULL,
    city                  TEXT           NOT NULL,
    description           TEXT,
    price                 BIGINT NOT NULL,
    duration_minutes      INT            NOT NULL,
    is_active             BOOLEAN        DEFAULT TRUE,
    is_auto_confirm       BOOLEAN        DEFAULT TRUE,
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_master_services_modtime
BEFORE UPDATE ON master_services
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- master_working_hours

CREATE TABLE IF NOT EXISTS master_working_hours (
    id          UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    master_id   UUID    NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    day_of_week INT     NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    start_time  TIME,
    end_time    TIME,
    is_day_off  BOOLEAN DEFAULT false,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_working_time_order CHECK (
        (is_day_off = true  AND start_time IS NULL     AND end_time IS NULL) OR
        (is_day_off = false AND start_time IS NOT NULL AND end_time IS NOT NULL AND start_time < end_time)
    ),
    UNIQUE(master_id, day_of_week)
);

CREATE INDEX IF NOT EXISTS idx_working_hours_master_id ON master_working_hours(master_id);

CREATE TRIGGER update_working_hours_modtime
BEFORE UPDATE ON master_working_hours
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- master_schedule_exceptions

CREATE TABLE IF NOT EXISTS master_schedule_exceptions (
    id             UUID    PRIMARY KEY DEFAULT uuid_generate_v4(),
    master_id      UUID    NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    exception_date DATE    NOT NULL,
    start_time     TIME,
    end_time       TIME,
    is_working     BOOLEAN NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_exception_time_order CHECK (
        (is_working = false AND start_time IS NULL     AND end_time IS NULL) OR
        (is_working = true  AND start_time IS NOT NULL AND end_time IS NOT NULL AND start_time < end_time)
    ),
    UNIQUE(master_id, exception_date)
);

CREATE INDEX IF NOT EXISTS idx_schedule_exceptions_master_date ON master_schedule_exceptions(master_id, exception_date);

CREATE TRIGGER update_schedule_exceptions_modtime
BEFORE UPDATE ON master_schedule_exceptions
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- appointments

CREATE TABLE IF NOT EXISTS appointments (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id       UUID        REFERENCES "user"(user_id) ON DELETE CASCADE,
    master_id       UUID        NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    service_id      UUID        REFERENCES master_services(id) ON DELETE RESTRICT,
    start_at        TIMESTAMP WITH TIME ZONE NOT NULL,
    end_at          TIMESTAMP WITH TIME ZONE NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    is_manual_block BOOLEAN     DEFAULT false,
    client_comment  TEXT,
    master_note     TEXT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_appointments_time_order CHECK (start_at < end_at),

    CONSTRAINT chk_appointments_status CHECK (
        status IN ('pending', 'confirmed', 'rejected', 'cancelled', 'completed')
    ),

    CONSTRAINT chk_manual_block CHECK (
        (is_manual_block = true  AND client_id IS NULL     AND service_id IS NULL) OR
        (is_manual_block = false AND client_id IS NOT NULL AND service_id IS NOT NULL)
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

-- master_portfolio_photos

CREATE TABLE IF NOT EXISTS master_portfolio_photos (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    master_id   UUID        NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    object_name TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_portfolio_photos_master_id ON master_portfolio_photos(master_id);

-- clients

CREATE TABLE IF NOT EXISTS clients (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID         NOT NULL UNIQUE REFERENCES "user"(user_id) ON DELETE CASCADE,
    first_name VARCHAR(255) NOT NULL,
    last_name  VARCHAR(255) NOT NULL DEFAULT '',
    phone      VARCHAR(20)   NOT NULL DEFAULT '' CHECK (phone = '' OR phone ~ '^\+[1-9][0-9]{6,14}$'),
    avatar_url VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_clients_modtime
BEFORE UPDATE ON clients
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
