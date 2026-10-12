CREATE TABLE embed_auth_tokens (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    secret_hash BYTEA NOT NULL,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
    expires_on TIMESTAMPTZ,
    used_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE INDEX embed_auth_tokens_project_email_idx ON embed_auth_tokens (project_id, email);