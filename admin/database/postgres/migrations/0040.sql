ALTER TABLE org_invites ALTER COLUMN org_role_id DROP NOT NULL;

ALTER TABLE org_invites ADD COLUMN usergroup_id UUID REFERENCES usergroups (id) ON DELETE CASCADE;

ALTER TABLE org_invites ADD CONSTRAINT org_role_id_or_usergroup_id_check
    CHECK (
        (org_role_id IS NOT NULL AND usergroup_id IS NULL) OR
        (org_role_id IS NULL AND usergroup_id IS NOT NULL)
    );
