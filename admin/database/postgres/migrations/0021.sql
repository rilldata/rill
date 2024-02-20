ALTER TABLE deployments
ADD COLUMN provisioner TEXT NOT NULL DEFAULT 'static' ,
ADD COLUMN provision_id TEXT NOT NULL DEFAULT '',
ADD COLUMN runtime_version TEXT NOT NULL DEFAULT '';

ALTER TABLE projects RENAME COLUMN region TO provisioner;
ALTER TABLE projects ADD COLUMN prod_version TEXT DEFAULT 'latest' NOT NULL;
