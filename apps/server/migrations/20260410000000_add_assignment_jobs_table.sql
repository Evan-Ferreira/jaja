-- +goose Up
CREATE TABLE assignment_jobs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id       UUID NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    anthropic_batch_id  TEXT,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    result              TEXT,
    error_message       TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assignment_jobs_status ON assignment_jobs(status);
CREATE INDEX idx_assignment_jobs_anthropic_batch_id ON assignment_jobs(anthropic_batch_id);

-- +goose Down
DROP TABLE assignment_jobs;
