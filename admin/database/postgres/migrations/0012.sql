ALTER TABLE user_auth_tokens ADD COLUMN used_on TIMESTAMPTZ DEFAULT now() NOT NULL;
ALTER TABLE service_auth_tokens ADD COLUMN used_on TIMESTAMPTZ DEFAULT now() NOT NULL;
ALTER TABLE users ADD COLUMN active_on TIMESTAMPTZ DEFAULT now() NOT NULL;

UPDATE user_auth_tokens SET used_on = created_on;
UPDATE service_auth_tokens SET used_on = created_on;
UPDATE users SET active_on = updated_on;
