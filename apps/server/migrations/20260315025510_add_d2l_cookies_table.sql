-- +goose Up
create table d2l_cookies (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    user_id         UUID NOT NULL REFERENCES users(id),
    clck         TEXT,  
    clsk       TEXT,
    d2l_same_site_canary_a TEXT,
    d2l_same_site_canary_b TEXT,
    d2l_secure_session_val TEXT,
    d2l_session_val TEXT
);

-- +goose Down
DROP TABLE d2l_cookies;