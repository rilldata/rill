ALTER TABLE users_projects_roles ADD COLUMN resources JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE usergroups_projects_roles ADD COLUMN resources JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE project_invites ADD COLUMN resources JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE users_projects_roles ADD COLUMN restrict_resources BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE usergroups_projects_roles ADD COLUMN restrict_resources BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE project_invites ADD COLUMN restrict_resources BOOLEAN NOT NULL DEFAULT FALSE;