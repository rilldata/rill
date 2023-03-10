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

ALTER TABLE projects
ADD COLUMN production_deployment_id UUID REFERENCES deployments ON DELETE SET NULL,
ADD COLUMN production_slots INTEGER NOT NULL DEFAULT 0;

CREATE UNIQUE INDEX projects_deployment_id_idx ON projects (production_deployment_id);
