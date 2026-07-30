ALTER TABLE orgs ADD COLUMN billing_portal_admin TEXT NOT NULL DEFAULT '';

UPDATE orgs SET billing_portal_admin = billing_email;
