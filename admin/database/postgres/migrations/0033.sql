CREATE TABLE project_access_request (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    email TEXT NOT NULL,
    project_id UUID REFERENCES projects (id) ON DELETE CASCADE,
    project_role_id UUID REFERENCES project_roles (id) ON DELETE CASCADE,
    created_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX project_access_request_email_project_idx ON project_access_request (lower(email), project_id);
