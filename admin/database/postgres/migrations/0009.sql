CREATE TABLE org_roles (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    name TEXT NOT NULL,
    read_org boolean,
    manage_org boolean,
    read_projects boolean,
    create_projects boolean,
    manage_projects boolean,
    read_org_members boolean,
    manage_org_members boolean
);

CREATE UNIQUE INDEX org_role_name_idx ON org_roles (name);

CREATE TABLE project_roles (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    name TEXT NOT NULL,
    read_project boolean,
    manage_project boolean,
    read_prod_branch boolean,
    manage_prod_branch boolean,
    read_dev_branches boolean,
    manage_dev_branches boolean,
    read_project_members boolean,
    manage_project_members boolean
);

CREATE UNIQUE INDEX project_role_name_idx ON project_roles (name);

INSERT INTO org_roles (id, name, read_org, manage_org, read_projects, create_projects, manage_projects, read_org_members, manage_org_members)
VALUES
    ('12345678-0000-0000-0000-000000000011', 'admin', true, true, true, true, true, true, true),
    ('12345678-0000-0000-0000-000000000012', 'collaborator', true, false, true, true, false, true, false),
    ('12345678-0000-0000-0000-000000000013', 'reader', true, false, true, false, false, true, false);

INSERT INTO project_roles (id, name, read_project, manage_project, read_prod_branch, manage_prod_branch, read_dev_branches, manage_dev_branches, read_project_members, manage_project_members)
VALUES
    ('12345678-0000-0000-0000-000000000021', 'admin', true, true, true, true, true, true, true, true),
    ('12345678-0000-0000-0000-000000000022', 'collaborator', true, false, true, false, true, true, true, false),
    ('12345678-0000-0000-0000-000000000023', 'reader', true, false, true, false, true, false, false, false);

CREATE TABLE users_orgs_roles (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    org_role_id UUID NOT NULL REFERENCES org_roles (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, org_id)
);

CREATE INDEX users_orgs_roles_org_user_idx ON users_orgs_roles (org_id, user_id);

CREATE TABLE users_projects_roles (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    project_role_id UUID NOT NULL REFERENCES project_roles (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, project_id)
);

CREATE INDEX users_projects_roles_project_user_idx ON users_projects_roles (project_id, user_id);

-- user group related table operations
CREATE TABLE usergroups (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    display_name TEXT NOT NULL
);

-- create all users group for existing organizations
INSERT INTO usergroups (org_id, display_name) SELECT id, 'all_users' FROM organizations;

-- add all_usergroup_id column to the organizations table referencing usergroups table
ALTER TABLE organizations ADD COLUMN all_usergroup_id UUID REFERENCES usergroups (id) ON DELETE RESTRICT;

-- insert all_usergroup_id into the organizations table for existing organizations
UPDATE organizations SET all_usergroup_id = (SELECT id FROM usergroups WHERE org_id = organizations.id);

CREATE TABLE users_usergroups (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    usergroup_id UUID NOT NULL REFERENCES usergroups (id) ON DELETE CASCADE
);

-- insert all users into all_users_group of corresponding organizations
WITH user_org AS (
    SELECT users.id as user_id, users_orgs_roles.org_id as org_id FROM users JOIN users_orgs_roles ON users_orgs_roles.user_id = users.id
)
INSERT INTO users_usergroups (user_id, usergroup_id)
SELECT user_org.user_id, org.all_usergroup_id FROM user_org JOIN organizations org ON org.id = user_org.org_id;

CREATE TABLE usergroups_orgs_roles (
    usergroup_id UUID NOT NULL REFERENCES usergroups (id) ON DELETE CASCADE,
    org_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    org_role_id UUID NOT NULL REFERENCES org_roles (id) ON DELETE CASCADE
);

CREATE TABLE usergroups_projects_roles (
    usergroup_id UUID NOT NULL REFERENCES usergroups (id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    project_role_id UUID NOT NULL REFERENCES project_roles (id) ON DELETE CASCADE
);

-- add all all_user_groups to projects with 'project collaborator' role
INSERT INTO usergroups_projects_roles (usergroup_id, project_id, project_role_id)
SELECT o.all_usergroup_id, p.id, '12345678-0000-0000-0000-000000000022'
FROM projects p JOIN organizations o ON p.organization_id = o.id;