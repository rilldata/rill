ALTER TABLE magic_auth_tokens ADD COLUMN secret TEXT NOT NULL DEFAULT '';
ALTER TABLE magic_auth_tokens ADD COLUMN secret_encryption_key_id TEXT NOT NULL DEFAULT '';
