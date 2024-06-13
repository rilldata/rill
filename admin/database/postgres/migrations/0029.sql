CREATE TABLE magic_auth_tokens (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    secret_hash BYTEA NOT NULL,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
    expires_on TIMESTAMPTZ,
    used_on TIMESTAMPTZ DEFAULT now() NOT NULL,
    dashboard TEXT NOT NULL,
    filter_json TEXT NOT NULL,
    exclude_fields TEXT[] NOT NULL
);

CREATE INDEX deployment_auth_tokens_deployment_idx ON deployment_auth_tokens (deployment_id);
