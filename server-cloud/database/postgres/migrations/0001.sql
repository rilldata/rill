CREATE EXTENSION "uuid-ossp";

CREATE TABLE organizations1 (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX organizations_name_idx ON organizations1 (name);

CREATE TABLE projects1 (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	organization_id UUID NOT NULL REFERENCES organizations1 (id) ON DELETE RESTRICT,
	name TEXT NOT NULL,
	description TEXT,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX projects_name_idx ON projects1 (organization_id, name);
