-- +goose Up
CREATE TABLE courses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    d2l_id          INTEGER NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    code            TEXT NOT NULL,
    description     TEXT,
    syllabus_s3_key TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE assignments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id         UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    d2l_id            INTEGER NOT NULL,
    name              TEXT NOT NULL,
    description       TEXT,
    instructions_text TEXT,
    due_date          TIMESTAMPTZ,
    score_out_of      DOUBLE PRECISION,
    is_hidden         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (course_id, d2l_id)
);

-- +goose Down
DROP TABLE assignments;
DROP TABLE courses;
