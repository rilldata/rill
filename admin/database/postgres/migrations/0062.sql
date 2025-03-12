-- Add default_project_role_id to table orgs and populate it with the 'viewer' role for all existing orgs.
-- The field should be nullable to indicate that no default project role should be applied.
ALTER TABLE orgs ADD COLUMN default_project_role_id UUID REFERENCES project_roles(id);
UPDATE orgs SET default_project_role_id = (SELECT id FROM project_roles WHERE name = 'viewer');
