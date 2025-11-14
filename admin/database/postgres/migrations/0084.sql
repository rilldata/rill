ALTER TABLE user_auth_tokens ADD COLUMN refresh BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE auth_clients ADD COLUMN scope TEXT NOT NULL DEFAULT '';
ALTER TABLE auth_clients ADD COLUMN used_on TIMESTAMPTZ NOT NULL DEFAULT now();

UPDATE auth_clients SET scope = 'long_lived_access_token';
