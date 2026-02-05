ALTER TABLE user_auth_tokens ADD COLUMN refresh BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE auth_clients ADD COLUMN scope TEXT NOT NULL DEFAULT '';
-- For Rill clients allow issuing long-lived access tokens
UPDATE auth_clients SET scope = 'long_lived_access_token';

ALTER TABLE auth_clients ADD COLUMN used_on TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE auth_clients ADD COLUMN grant_types TEXT[] NOT NULL DEFAULT ARRAY['authorization_code'];
-- For Rill CLI client
UPDATE auth_clients SET grant_types = ARRAY['urn:ietf:params:oauth:grant-type:device_code'] WHERE id = '12345678-0000-0000-0000-000000000002';
