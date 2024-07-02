ALTER TABLE project_roles ADD create_magic_auth_tokens BOOLEAN DEFAULT false NOT NULL;
UPDATE project_roles SET create_magic_auth_tokens = read_project;
