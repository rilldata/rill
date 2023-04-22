CREATE EXTENSION "uuid-ossp";

CREATE TABLE orgs (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT NOT NULL,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX orgs_name_idx ON orgs (lower(name));

CREATE TABLE projects (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	org_id UUID NOT NULL REFERENCES orgs (id) ON DELETE RESTRICT,
	name TEXT NOT NULL,
	description TEXT NOT NULL,
	public BOOLEAN NOT NULL,
	region TEXT NOT NULL,
	github_url TEXT,
	github_installation_id BIGINT,
	prod_branch TEXT NOT NULL,
	prod_variables JSONB DEFAULT '{}'::jsonb NOT NULL,
	prod_olap_driver TEXT NOT NULL,
	prod_olap_dsn TEXT NOT NULL,
	prod_slots INTEGER NOT NULL,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX projects_name_idx ON projects (org_id, lower(name));
CREATE INDEX projects_github_url_idx ON projects (lower(github_url), org_id);

CREATE TABLE deployments (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	project_id UUID NOT NULL REFERENCES projects (id) ON DELETE RESTRICT,
	slots INTEGER NOT NULL,
	branch TEXT NOT NULL,
	runtime_host TEXT NOT NULL,
	runtime_instance_id TEXT NOT NULL,
	runtime_audience TEXT NOT NULL,
	status INTEGER NOT NULL,
	logs TEXT NOT NULL,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE INDEX deployments_project_id_idx ON deployments (project_id);

ALTER TABLE projects ADD COLUMN prod_deployment_id UUID REFERENCES deployments ON DELETE SET NULL;
CREATE UNIQUE INDEX projects_deployment_id_idx ON projects (prod_deployment_id);
