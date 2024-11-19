CREATE TABLE provisioner_resources (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    deployment_id UUID NOT NULL REFERENCES deployments (id) ON DELETE RESTRICT,
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

CREATE UNIQUE INDEX provisioner_resources_deployment_id_type_name_idx ON provisioner_resources (deployment_id, "type", lower(name));

INSERT INTO provisioner_resources (
    id,
    deployment_id,
    "type",
    name,
    status,
    status_message,
    provisioner,
    args_json,
    state_json,
    config_json,
    created_on,
    updated_on
) SELECT
    uuid(provision_id),
    id as deployment_id,
    'runtime' as "type",
    '' as name,
    status,
    status_message,
    provisioner,
    jsonb_build_object('slots', slots, 'version', 'latest') as args_json,
    jsonb_build_object('slots', slots, 'version', runtime_version) as state_json,
    jsonb_build_object('host', runtime_host, 'audience', runtime_audience, 'cpu', slots, 'memory_gb', 4*slots, 'storage_bytes', 40*slots) as config_json,
    created_on,
    updated_on
FROM deployments;

CREATE TABLE static_runtime_assignments (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    -- This could be a foreign key to provisioner_resources.id, but we don't enforce it to avoid tight coupling with the provisioner's internal data model.
    resource_id UUID NOT NULL,
    host TEXT NOT NULL,
    slots INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX static_runtime_assignments_resource_id_idx ON static_runtime_assignments (resource_id);

INSERT INTO static_runtime_assignments (resource_id, host, slots)
SELECT uuid(d.provision_id), d.runtime_host, d.slots FROM deployments d WHERE d.provisioner = 'static';

ALTER TABLE deployments
DROP COLUMN slots,
DROP COLUMN provisioner,
DROP COLUMN provision_id,
DROP COLUMN runtime_version;
