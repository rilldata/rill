CREATE TABLE service_orgs_roles (
    service_id UUID PRIMARY KEY NOT NULL REFERENCES service (id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES orgs (id) ON DELETE CASCADE,
    org_role_id UUID NOT NULL REFERENCES org_roles (id) ON DELETE CASCADE
);

CREATE INDEX service_orgs_roles_org_service_idx ON service_orgs_roles (org_id, service_id);

INSERT INTO service_orgs_roles (service_id, org_id, org_role_id)
SELECT s.id, o.id, r.id
FROM service s
JOIN orgs o ON o.id = s.org_id
JOIN org_roles r ON r.name = 'admin';

CREATE TABLE service_projects_roles (
    service_id UUID NOT NULL REFERENCES service (id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    project_role_id UUID NOT NULL REFERENCES project_roles (id) ON DELETE CASCADE,
    PRIMARY KEY (service_id, project_id)
);

CREATE INDEX service_projects_roles_project_service_idx ON service_projects_roles (project_id, service_id);

ALTER TABLE service ADD COLUMN attributes JSONB NOT NULL DEFAULT '{}'::jsonb;