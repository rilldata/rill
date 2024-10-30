ALTER TABLE org_invites ADD COLUMN usergroup_ids TEXT[] NOT NULL DEFAULT '{}'::TEXT[];
