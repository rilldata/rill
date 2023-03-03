CREATE TABLE device_code_auth (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    device_code TEXT NOT NULL,
    user_code TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    approval_state INTEGER NOT NULL,
    client_id TEXT NOT NULL,
    user_id TEXT NOT NULL
);

CREATE UNIQUE INDEX device_code_auth_device_code ON device_code_auth (device_code);
CREATE INDEX device_code_auth_user_code ON device_code_auth (user_code);