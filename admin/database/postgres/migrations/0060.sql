-- This is an idempotent version of migration 0057.sql.
-- It's needed because we accidentally sent out a cherry-picked patch of 0058.sql without 0057.sql.
ALTER TABLE orgs ADD COLUMN IF NOT EXISTS favicon_asset_id UUID REFERENCES assets(id) ON DELETE SET NULL;
