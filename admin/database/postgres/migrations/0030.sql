ALTER TABLE projects ADD COLUMN upload_path TEXT;

CREATE TABLE assets (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL,
    path TEXT NOT NULL,
    created_on TIMESTAMPTZ NOT NULL DEFAULT now()
);