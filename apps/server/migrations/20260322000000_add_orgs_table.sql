-- +goose Up
CREATE TABLE orgs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_name TEXT NOT NULL,
    d2l_base_url TEXT NOT NULL,
    le_version TEXT NOT NULL,
    lp_version TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE orgs;
