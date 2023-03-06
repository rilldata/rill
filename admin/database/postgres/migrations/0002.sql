CREATE TABLE users (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	email TEXT NOT NULL,
	display_name TEXT NOT NULL,
	photo_url TEXT,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX users_email_idx ON users (lower(email));

CREATE TABLE auth_clients (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	display_name TEXT NOT NULL,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE user_auth_tokens (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	secret_hash BYTEA NOT NULL,
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	display_name TEXT,
	auth_client_id UUID REFERENCES auth_clients (id) ON DELETE CASCADE,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE INDEX user_auth_tokens_user_idx ON user_auth_tokens (user_id);

-- Hard-coded first-party auth clients
INSERT INTO auth_clients (id, display_name)
VALUES
	('12345678-0000-0000-0000-000000000001', 'Rill Web'),
	('12345678-0000-0000-0000-000000000002', 'Rill CLI');
