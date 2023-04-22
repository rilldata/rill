CREATE TABLE users (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	email TEXT NOT NULL,
	display_name TEXT NOT NULL,
	photo_url TEXT,
	github_username TEXT NOT NULL,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX users_email_idx ON users (lower(email));

CREATE TABLE usergroups (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	org_id UUID NOT NULL REFERENCES orgs (id) ON DELETE CASCADE,
	name TEXT NOT NULL
);

CREATE UNIQUE INDEX usergroups_name_idx ON usergroups (org_id, lower(name));

ALTER TABLE orgs ADD COLUMN all_usergroup_id UUID REFERENCES usergroups (id) ON DELETE RESTRICT;

CREATE TABLE usergroups_users (
	usergroup_id UUID NOT NULL REFERENCES usergroups (id) ON DELETE CASCADE,
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	PRIMARY KEY (usergroup_id, user_id)
);

CREATE INDEX usergroups_users_user_usergroup_idx ON usergroups_users (user_id, usergroup_id);

CREATE TABLE auth_clients (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	display_name TEXT NOT NULL,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

-- Hard-coded first-party auth clients
INSERT INTO auth_clients (id, display_name)
VALUES
	('12345678-0000-0000-0000-000000000001', 'Rill Web'),
	('12345678-0000-0000-0000-000000000002', 'Rill CLI');

CREATE TABLE user_auth_tokens (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	secret_hash BYTEA NOT NULL,
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	display_name TEXT NOT NULL,
	auth_client_id UUID REFERENCES auth_clients (id) ON DELETE CASCADE,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE INDEX user_auth_tokens_user_idx ON user_auth_tokens (user_id);

CREATE TABLE device_auth_codes (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	device_code TEXT NOT NULL,
	user_code TEXT NOT NULL,
	expires_on TIMESTAMPTZ NOT NULL,
	approval_state INTEGER NOT NULL,
	client_id UUID NOT NULL REFERENCES auth_clients (id),
	user_id UUID REFERENCES users (id) ON DELETE CASCADE,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX device_auth_codes_device_code_idx ON device_auth_codes (device_code);
CREATE INDEX device_auth_codes_user_code_idx ON device_auth_codes (user_code);
