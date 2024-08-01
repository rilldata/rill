ALTER TABLE orgs ADD COLUMN billing_email TEXT NOT NULL DEFAULT '';
-- update existing orgs billing_email with the oldest admin email
UPDATE orgs SET billing_email = (SELECT email FROM users WHERE id IN (SELECT user_id FROM users_orgs_roles WHERE org_id = orgs.id AND org_role_id = (SELECT id FROM org_roles WHERE name = 'admin')) ORDER BY created_on ASC LIMIT 1);
