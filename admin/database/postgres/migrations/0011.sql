CREATE TABLE service (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	org_id UUID NOT NULL REFERENCES orgs (id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX service_name_idx ON service (org_id, lower(name));

CREATE TABLE service_auth_tokens (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	secret_hash BYTEA NOT NULL,
	service_id UUID NOT NULL REFERENCES service (id) ON DELETE CASCADE,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	expires_on TIMESTAMPTZ
);

CREATE INDEX service_auth_tokens_service_idx ON service_auth_tokens (service_id);
