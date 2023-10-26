ALTER TABLE project_roles ADD create_reports BOOLEAN NOT NULL DEFAULT false;
UPDATE project_roles SET create_reports = read_prod;

ALTER TABLE project_roles ADD manage_reports BOOLEAN NOT NULL DEFAULT false;
UPDATE project_roles SET manage_reports = manage_prod;
