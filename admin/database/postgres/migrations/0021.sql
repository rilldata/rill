ALTER TABLE deployments
ADD COLUMN provisioner TEXT DEFAULT 'static' NOT NULL,
ADD COLUMN provision_id TEXT,
ADD COLUMN runtime_version TEXT;

ALTER TABLE projects ADD COLUMN prod_runtime_version TEXT DEFAULT 'latest' NOT NULL;
