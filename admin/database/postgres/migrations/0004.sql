CREATE TABLE user_org_invites (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    email TEXT NOT NULL,
    invited_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    org_id UUID NOT NULL REFERENCES orgs (id) ON DELETE CASCADE,
    org_role_id UUID NOT NULL REFERENCES org_roles (id) ON DELETE CASCADE,
    created_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX user_org_invites_email_org_id_idx ON user_org_invites (lower(email), org_id);

CREATE TABLE user_project_invites (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    email TEXT NOT NULL,
    invited_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects (id) ON DELETE CASCADE,
    project_role_id UUID REFERENCES project_roles (id) ON DELETE CASCADE,
    created_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX user_project_invites_email_project_id_idx ON user_project_invites (lower(email), project_id);
