ALTER TABLE orgs ADD COLUMN custom_domain TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX orgs_custom_domain_idx ON orgs (lower(custom_domain)) WHERE custom_domain <> '';
