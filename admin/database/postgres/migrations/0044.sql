ALTER TABLE magic_auth_tokens ADD COLUMN secret BYTEA NOT NULL DEFAULT ''::BYTEA;
ALTER TABLE magic_auth_tokens ADD COLUMN secret_encryption_key_id TEXT NOT NULL DEFAULT '';
