-- +goose Up
CREATE TABLE jobs (
    id              TEXT PRIMARY KEY,
    queue           TEXT NOT NULL DEFAULT 'default',
    type            TEXT NOT NULL,
    payload         JSONB,
    state           TEXT NOT NULL,
    max_retry       INT NOT NULL DEFAULT 25,
    retried         INT NOT NULL DEFAULT 0,
    last_err        TEXT,
    last_failed_at  TIMESTAMPTZ,
    result          BYTEA,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE jobs;
