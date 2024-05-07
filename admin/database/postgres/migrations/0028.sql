-- Hard-coded first-party auth clients
INSERT INTO auth_clients (id, display_name)
VALUES ('12345678-0000-0000-0000-000000000004', 'Web Local');

-- Table for storing authorization codes for PKCE auth flow
CREATE TABLE authorization_codes (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    code TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES auth_clients(id) ON DELETE CASCADE,
    redirect_uri TEXT NOT NULL,
    code_challenge TEXT NOT NULL,
    code_challenge_method TEXT NOT NULL,
    expires_on TIMESTAMP NOT NULL,
    created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

-- create index on code column
CREATE UNIQUE INDEX authorization_codes_code_idx ON authorization_codes(code);