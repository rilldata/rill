CREATE TABLE device_code_auth (
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

CREATE UNIQUE INDEX device_code_auth_device_code ON device_code_auth (device_code);
CREATE INDEX device_code_auth_user_code ON device_code_auth (user_code);