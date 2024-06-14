CREATE TABLE magic_auth_tokens (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    secret_hash BYTEA NOT NULL,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
    expires_on TIMESTAMPTZ,
    used_on TIMESTAMPTZ DEFAULT now() NOT NULL,
    created_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    attributes JSONB DEFAULT '{}'::JSONB NOT NULL,
    metrics_view TEXT NOT NULL,
    metrics_view_filter_json TEXT NOT NULL,
    metrics_view_fields TEXT[] NOT NULL
);

CREATE INDEX magic_auth_tokens_project_id_idx ON magic_auth_tokens (project_id);
CREATE INDEX magic_auth_tokens_created_by_user_id_idx ON magic_auth_tokens (created_by_user_id) WHERE created_by_user_id IS NOT NULL;

ALTER TABLE project_roles ADD create_magic_auth_tokens BOOLEAN DEFAULT false NOT NULL;
UPDATE project_roles SET create_magic_auth_tokens = manage_project_members;

ALTER TABLE project_roles ADD manage_magic_auth_tokens BOOLEAN DEFAULT false NOT NULL;
UPDATE project_roles SET manage_magic_auth_tokens = manage_project_members;
