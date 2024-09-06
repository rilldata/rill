ALTER TABLE magic_auth_tokens ADD COLUMN token_str TEXT NOT NULL DEFAULT '';
ALTER TABLE magic_auth_tokens ADD COLUMN token_str_encryption_key_id TEXT NOT NULL DEFAULT '';
