-- Add EncryptedSecret column to the magic_auth_tokens table
ALTER TABLE magic_auth_tokens
ADD COLUMN encrypted_secret BYTEA NOT NULL;
