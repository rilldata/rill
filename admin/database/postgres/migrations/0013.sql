DROP INDEX organizations_name_idx;
CREATE UNIQUE INDEX organizations_name_idx ON organizations (lower(name));

DROP INDEX projects_name_idx;
CREATE UNIQUE INDEX projects_name_idx ON projects (org_id, lower(name));

DROP INDEX org_roles_name_idx;
CREATE UNIQUE INDEX org_roles_name_idx ON org_roles (lower(name));

DROP INDEX project_roles_name_idx;
CREATE UNIQUE INDEX project_roles_name_idx ON project_roles (lower(name));

CREATE UNIQUE INDEX usergroups_name_idx ON usergroups (org_id, lower(name));
