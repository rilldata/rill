-- Adds JSONB column 'attributes' for storing additional metadata
ALTER TABLE users_orgs_roles ADD COLUMN attributes JSONB DEFAULT '{}'::JSONB NOT NULL;
