CREATE TABLE provisioner_resources (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE SET NULL,
    "type" TEXT NOT NULL,
    name TEXT NOT NULL,
    status INTEGER NOT NULL,
    status_message TEXT NOT NULL DEFAULT '',
    provisioner TEXT NOT NULL,
    args_json JSONB NOT NULL,
    state_json JSONB NOT NULL,
    config_json JSONB NOT NULL,
    created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX provisioner_resources_project_id_type_name_idx ON provisioner_resources (project_id, "type", lower(name));

