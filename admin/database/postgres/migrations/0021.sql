ALTER TABLE deployments
ADD COLUMN provisioner TEXT DEFAULT 'static' NOT NULL,
ADD COLUMN provision_id TEXT,
ADD COLUMN runtime_version TEXT;

ALTER TABLE projects RENAME COLUMN region TO provisioner;
ALTER TABLE projects ADD COLUMN prod_version TEXT DEFAULT 'latest' NOT NULL;
