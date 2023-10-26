ALTER TABLE project_roles ADD COLUMN create_reports BOOLEAN NOT NULL DEFAULT false;
UPDATE TABLE project_roles SET create_reports = read_prod;

ALTER TABLE project_roles ADD COLUMN manage_reports BOOLEAN NOT NULL DEFAULT false;
UPDATE TABLE project_roles SET manage_reports = manage_prod;
