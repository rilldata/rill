CREATE TABLE org_invites (
	id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
	email TEXT NOT NULL,
	org_id UUID NOT NULL REFERENCES orgs (id) ON DELETE CASCADE,
	org_role_id UUID NOT NULL REFERENCES org_roles (id) ON DELETE CASCADE,
	invited_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
	created_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX org_invites_email_org_idx ON org_invites (lower(email), org_id);

CREATE TABLE project_invites (
	id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
	email TEXT NOT NULL,
	project_id UUID REFERENCES projects (id) ON DELETE CASCADE,
	project_role_id UUID REFERENCES project_roles (id) ON DELETE CASCADE,
	invited_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
	created_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX project_invites_email_project_idx ON project_invites (lower(email), project_id);
