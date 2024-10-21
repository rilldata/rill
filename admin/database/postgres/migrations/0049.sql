CREATE TABLE project_variables (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    value BYTEA NOT NULL,
    -- Encryption key ID if the value is encrypted
    value_encryption_key_id TEXT NOT NULL DEFAULT '',
    -- Environment it belongs to ("production" or "development").
    -- If empty, then it should be used as the fallback for all environments.
    environment TEXT NOT NULL DEFAULT '',
    -- The user who most recently edited the variable
    updated_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

-- And a unique index on (project_id, environment, lower(name))
CREATE UNIQUE INDEX project_variables_project_id_environment_name_idx ON project_variables (project_id, environment, lower(name));

-- Migration
WITH project_data AS (
  SELECT
    id AS project_id,
    prod_variables,
    prod_variables_encryption_key_id
  FROM
    projects
  WHERE prod_variables::JSON::TEXT NOT LIKE 'null'
)
INSERT INTO project_variables (
  project_id,
  name,
  value,
  value_encryption_key_id
)
SELECT
  project_data.project_id,
  key AS name,
  convert_to(value, 'UTF8')::bytea AS value,
  project_data.prod_variables_encryption_key_id AS value_encryption_key_id
FROM
  project_data,
  jsonb_each_text(project_data.prod_variables);

-- Drop the old column
ALTER TABLE projects DROP COLUMN prod_variables;
ALTER TABLE projects DROP COLUMN prod_variables_encryption_key_id;