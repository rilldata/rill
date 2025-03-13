-- Introduce the notion of managed usergroups, which are usergroups that are created and managed by the system.
-- They will not be directly editable by users.
ALTER TABLE usergroups ADD COLUMN managed BOOLEAN NOT NULL DEFAULT false;

-- Now we don't need all_usergroup_id on orgs anymore.
ALTER TABLE orgs DROP COLUMN all_usergroup_id;

-- Migrate the `all-users` group to be a managed group.
UPDATE usergroups SET managed = true WHERE name = 'all-users';

-- Migrate the "collaborator" role to be called "editor".
UPDATE org_roles SET name = 'editor' WHERE name = 'collaborator';
UPDATE project_roles SET name = 'editor' WHERE name = 'collaborator';

-- Add a flag to indicate if an org role is considered a guest role.
ALTER TABLE org_roles ADD guest BOOLEAN NOT NULL DEFAULT false;

-- Create an org role 'guest' that is similar to 'viewer' but marked as a guest role.
INSERT INTO org_roles (name, guest, read_org, manage_org, read_projects, create_projects, manage_projects, read_org_members, manage_org_members)
VALUES ('guest', true, true, false, true, false, false, false, false);

-- Add all project-level members who are not already org-level members as org-level members with the 'guest' role.
INSERT INTO users_orgs_roles (user_id, org_id, org_role_id)
SELECT
    upr.user_id AS user_id,
    p.org_id AS org_id,
    (SELECT ors.id FROM org_roles ors WHERE ors.name = 'guest') AS org_role_id
FROM users_projects_roles upr
JOIN projects p ON upr.project_id = p.id
WHERE upr.user_id NOT IN (SELECT uor.user_id FROM users_orgs_roles uor WHERE uor.org_id = p.org_id);

-- We want to have three managed user groups: all-users, all-members, and all-guests.
-- Firstly, we rename the `all-users` group to `all-members` since it doesn't yet contain the guest users.
-- This also preserves the existing project permissions that are based on the `all-users` group.
UPDATE usergroups SET name = 'all-members' WHERE name = 'all-users';

-- Create a new managed group `all-users` and add all org-level members to it, including guests.
INSERT INTO usergroups (org_id, name, managed)
SELECT id as org_id, 'all-users' AS name, true AS managed FROM orgs;

INSERT INTO usergroups_users (usergroup_id, user_id)
SELECT
    ug.id AS usergroup_id,
    uor.user_id AS user_id
FROM usergroups ug
JOIN users_orgs_roles uor ON ug.org_id = uor.org_id
WHERE ug.name = 'all-users';

-- Create a new managed group `all-guests` and add all guest org-level members to it.
INSERT INTO usergroups (org_id, name, managed)
SELECT id as org_id, 'all-guests' AS name, true AS managed FROM orgs;

INSERT INTO usergroups_users (usergroup_id, user_id)
SELECT
    ug.id AS usergroup_id,
    uor.user_id AS user_id
FROM usergroups ug
JOIN users_orgs_roles uor ON ug.org_id = uor.org_id
WHERE ug.name = 'all-guests' AND uor.org_role_id = (SELECT ors.id FROM org_roles ors WHERE ors.name = 'guest');
