ALTER TABLE orgs
  ADD COLUMN logo_dark_asset_id UUID REFERENCES assets(id) ON DELETE SET NULL;

