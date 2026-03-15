-- +goose Up
CREATE TABLE d2l_local_storages (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id),
    d2l_fetch_tokens     TEXT,
    session_expired      TEXT,
    session_last_accessed TEXT,
    session_user_id      TEXT,
    xsrf_hit_code_seed   TEXT,
    xsrf_token           TEXT,
    pdfjs_history        TEXT,                   
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE d2l_local_storages;