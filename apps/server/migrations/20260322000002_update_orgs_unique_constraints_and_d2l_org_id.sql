-- +goose Up
ALTER TABLE orgs ADD COLUMN d2l_org_id TEXT NOT NULL DEFAULT '';
ALTER TABLE orgs ALTER COLUMN d2l_org_id DROP DEFAULT;
ALTER TABLE orgs ADD CONSTRAINT orgs_org_name_key UNIQUE (org_name);
ALTER TABLE orgs ADD CONSTRAINT orgs_d2l_org_id_key UNIQUE (d2l_org_id);
ALTER TABLE orgs ADD CONSTRAINT orgs_d2l_base_url_key UNIQUE (d2l_base_url);

-- +goose Down
ALTER TABLE orgs DROP CONSTRAINT orgs_org_name_key;
ALTER TABLE orgs DROP CONSTRAINT orgs_d2l_org_id_key;
ALTER TABLE orgs DROP CONSTRAINT orgs_d2l_base_url_key;
ALTER TABLE orgs DROP COLUMN d2l_org_id;
