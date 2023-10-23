CREATE TABLE deployment_auth_tokens (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	secret_hash BYTEA NOT NULL,
	deployment_id UUID NOT NULL REFERENCES deployments (id) ON DELETE CASCADE,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	used_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	expires_on TIMESTAMPTZ
);

CREATE INDEX deployment_auth_tokens_deployment_idx ON deployment_auth_tokens (deployment_id);
