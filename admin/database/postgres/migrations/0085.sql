ALTER TABLE users_projects_roles ADD COLUMN resources JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE usergroups_projects_roles ADD COLUMN resources JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE project_invites ADD COLUMN resources JSONB NOT NULL DEFAULT '[]'::jsonb;
