CREATE TABLE master_portfolio_photos (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    master_id   UUID        NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    object_name TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_portfolio_photos_master_id ON master_portfolio_photos(master_id);
