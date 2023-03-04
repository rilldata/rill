CREATE TABLE users (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    email TEXT NOT NULL,
	display_name TEXT NOT NULL,
    photo_url TEXT NOT NULL,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX users_email_idx ON users (lower(email));

CREATE TABLE user_auth_tokens (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    secret_hash BYTEA NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id),
    display_name TEXT,
    oauth_client_id UUID,
    created_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE INDEX user_auth_tokens_user_idx ON user_auth_tokens (user_id);
