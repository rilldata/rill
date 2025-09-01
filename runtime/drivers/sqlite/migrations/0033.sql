ALTER TABLE conversations ADD COLUMN app_context_type TEXT NOT NULL DEFAULT '';
ALTER TABLE conversations ADD COLUMN app_context_metadata_json TEXT NOT NULL DEFAULT '{}'; 