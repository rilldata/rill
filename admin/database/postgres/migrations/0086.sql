ALTER TABLE auth_clients ADD COLUMN redirect_uris TEXT[] NOT NULL DEFAULT '{}'::TEXT[];
