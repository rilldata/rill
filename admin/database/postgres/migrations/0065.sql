-- This migration ensures there's an org_invite for every project_invite.
-- This ensures that listing org-level invites covers all invites within the org,
-- and removing an org-level invite will remove all project-level invites for that email.

-- Add the org_invite_id column
ALTER TABLE project_invites ADD COLUMN org_invite_id UUID;

-- Add the foreign key
ALTER TABLE project_invites
ADD CONSTRAINT project_invites_org_invite_id_fkey
FOREIGN KEY (org_invite_id)
REFERENCES org_invites (id)
ON DELETE CASCADE;

-- Create org invites for project invites that don't have one
INSERT INTO org_invites (email, org_id, org_role_id, invited_by_user_id, created_on)
SELECT pi.email, p.org_id, (SELECT id FROM org_roles WHERE name = 'guest'), pi.invited_by_user_id, pi.created_on
FROM project_invites pi
JOIN projects p ON pi.project_id = p.id
WHERE NOT EXISTS (SELECT 1 FROM org_invites oi WHERE oi.email = pi.email AND oi.org_id = p.org_id);

-- Update the project invites with the org invite ids
UPDATE project_invites pi
SET org_invite_id = oi.id
FROM org_invites oi
JOIN projects p ON oi.org_id = p.org_id
WHERE pi.email = oi.email AND p.id = pi.project_id;

-- Add the not null constraint
ALTER TABLE project_invites ALTER COLUMN org_invite_id SET NOT NULL;
