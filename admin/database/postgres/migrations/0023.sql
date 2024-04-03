ALTER TABLE project_roles ADD create_alerts BOOLEAN NOT NULL DEFAULT false;
UPDATE project_roles SET create_alerts = read_prod;

ALTER TABLE project_roles ADD manage_alerts BOOLEAN NOT NULL DEFAULT false;
UPDATE project_roles SET manage_alerts = manage_prod;
