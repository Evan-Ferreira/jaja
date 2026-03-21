-- +goose Up
ALTER TABLE d2l_local_storages
    DROP COLUMN IF EXISTS d2l_fetch_tokens,
    ADD COLUMN fetch_access_token TEXT,
    ADD COLUMN fetch_expires_at   BIGINT;

-- +goose Down
ALTER TABLE d2l_local_storages
    DROP COLUMN IF EXISTS fetch_access_token,
    DROP COLUMN IF EXISTS fetch_expires_at,
    ADD COLUMN d2l_fetch_tokens TEXT;
