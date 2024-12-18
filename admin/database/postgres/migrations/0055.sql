ALTER TABLE project_roles ADD read_provisioner_resources BOOLEAN DEFAULT false NOT NULL;
UPDATE project_roles SET read_provisioner_resources = read_prod_status;

ALTER TABLE project_roles ADD manage_provisioner_resources BOOLEAN DEFAULT false NOT NULL;
UPDATE project_roles SET manage_provisioner_resources = manage_prod;
