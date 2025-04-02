-- Add a flag to indicate if an org role is considered an admin role.
ALTER TABLE org_roles ADD admin BOOLEAN NOT NULL DEFAULT false;
UPDATE org_roles SET admin = true WHERE manage_org;

-- Add a flag to indicate if a project role is considered an admin role.
ALTER TABLE project_roles ADD admin BOOLEAN NOT NULL DEFAULT false;
UPDATE project_roles SET admin = true WHERE manage_project;

-- Update the 'editor' org role to have manage_org_members permissions.
UPDATE org_roles SET manage_org_members = true WHERE name = 'editor';

-- Update the 'editor' project role to have manage_project_members permissions.
UPDATE project_roles SET manage_project_members = true WHERE name = 'editor';

-- Add a separate permission for managing admin roles.
ALTER TABLE org_roles ADD manage_org_admins BOOLEAN NOT NULL DEFAULT false;
UPDATE org_roles SET manage_org_admins = manage_org;

ALTER TABLE project_roles ADD manage_project_admins BOOLEAN NOT NULL DEFAULT false;
UPDATE project_roles SET manage_project_admins = manage_project;
