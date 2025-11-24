ALTER TABLE deployments ADD COLUMN desired_status INTEGER DEFAULT 0 NOT NULL;
ALTER TABLE deployments ADD COLUMN desired_status_updated_on TIMESTAMPTZ DEFAULT now() NOT NULL;

