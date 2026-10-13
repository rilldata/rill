ALTER TABLE magic_auth_tokens ADD COLUMN password_hash BYTEA;

CREATE TABLE magic_auth_token_sessions (
    id UUID PRIMARY KEY,
    secret_hash BYTEA NOT NULL,
    magic_auth_token_id UUID NOT NULL REFERENCES magic_auth_tokens (id) ON DELETE CASCADE,
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_on TIMESTAMPTZ NOT NULL
);

CREATE INDEX magic_auth_token_sessions_magic_auth_token_id_idx ON magic_auth_token_sessions (magic_auth_token_id);
