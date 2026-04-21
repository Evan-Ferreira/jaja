-- +goose Up
ALTER TABLE jobs
    ALTER COLUMN result TYPE jsonb
    USING (
        CASE
            WHEN result IS NULL THEN NULL
            ELSE convert_from(result, 'UTF8')::jsonb
        END
    );

-- +goose Down
ALTER TABLE jobs
    ALTER COLUMN result TYPE bytea
    USING (
        CASE
            WHEN result IS NULL THEN NULL
            ELSE convert_to(result::text, 'UTF8')
        END
    );
