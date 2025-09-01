-- Reconfigure `branch` to `environment` in virtual_files.
ALTER TABLE virtual_files RENAME COLUMN branch TO environment;

-- Rename the index
DROP INDEX virtual_files_project_id_branch_updated_on_idx;
CREATE INDEX virtual_files_project_id_environment_updated_on_idx ON virtual_files (project_id, environment, updated_on);

-- Currently the environment for all projects is hard-coded to "prod".
UPDATE virtual_files SET environment = 'prod';
