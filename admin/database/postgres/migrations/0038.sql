ALTER TABLE orgs ADD COLUMN billing_email TEXT NOT NULL DEFAULT '';
-- update existing orgs billing_email with the oldest admin email
UPDATE orgs SET billing_email = (SELECT email FROM users WHERE id IN (SELECT user_id FROM users_orgs_roles WHERE org_id = orgs.id AND org_role_id = (SELECT id FROM org_roles WHERE name = 'admin')) ORDER BY created_on ASC LIMIT 1);

-- update quota_storage_limit_bytes_per_deployment in orgs table from 5GB to 10GB
UPDATE orgs SET quota_storage_limit_bytes_per_deployment = 10737418240 WHERE quota_storage_limit_bytes_per_deployment = 5368709120;
