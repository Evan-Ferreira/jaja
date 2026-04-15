-- +goose Up
CREATE TABLE jobs (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload    JSONB,
    job_type   TEXT NOT NULL,
    status     TEXT NOT NULL,
    result     JSONB,
    error      TEXT
);

-- +goose Down
DROP TABLE jobs;
