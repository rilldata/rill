-- Add a flag to indicate if an org role is considered a guest role.
ALTER TABLE org_roles ADD guest BOOLEAN NOT NULL DEFAULT false;

-- Create an org role 'guest' that is similar to 'viewer' but marked as a guest role.
INSERT INTO org_roles (name, guest, read_org, manage_org, read_projects, create_projects, manage_projects, read_org_members, manage_org_members)
VALUES ('guest', true, true, false, true, false, false, false, false);

-- Add all project-level members who are not already org-level members as org-level members with the 'guest' role.
INSERT INTO users_orgs_roles (user_id, org_id, role_id)
SELECT
    upr.user_id AS user_id,
    p.org_id AS org_id,
    (SELECT ors.id FROM org_roles ors WHERE ors.name = 'guest') AS org_role_id
FROM users_projects_roles upr
JOIN projects p ON upr.project_id = p.id
WHERE upr.user_id NOT IN (SELECT uor.user_id FROM users_orgs_roles uor WHERE uor.org_id = p.org_id);

