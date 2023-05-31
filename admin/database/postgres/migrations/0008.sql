ALTER TABLE user_auth_tokens ADD COLUMN representing_user_id UUID REFERENCES users (id) ON DELETE CASCADE;
ALTER TABLE user_auth_tokens ADD COLUMN expires_on TIMESTAMPTZ;

-- Hard-coded first-party auth clients
INSERT INTO auth_clients (id, display_name)
VALUES ('12345678-0000-0000-0000-000000000003', 'Rill Support');
