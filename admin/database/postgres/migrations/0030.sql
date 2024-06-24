ALTER TABLE projects ADD COLUMN archive_asset_id UUID;

CREATE TABLE assets (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    path TEXT NOT NULL,
    owner_id UUID not null,
    created_on TIMESTAMPTZ NOT NULL DEFAULT now()
);