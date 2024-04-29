CREATE TABLE projects_autoinvite_domains (
    id UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    project_role_id UUID NOT NULL REFERENCES project_roles (id) ON DELETE CASCADE,
    domain TEXT NOT NULL,
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX projects_autoinvite_domains_domain_idx ON projects_autoinvite_domains (lower(domain));
CREATE UNIQUE INDEX projects_autoinvite_domains_project_id_domain_idx ON projects_autoinvite_domains (project_id, lower(domain));
