ALTER TABLE orgs ADD COLUMN billing_customer_id TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX orgs_billing_customer_id_idx ON orgs (billing_customer_id) WHERE billing_customer_id <> '';

ALTER TABLE orgs ADD COLUMN quota_num_users INTEGER NOT NULL DEFAULT -1; -- review: do we need quota on num admins as well
ALTER TABLE orgs ADD COLUMN quota_managed_data_bytes BIGINT NOT NULL DEFAULT -1;
UPDATE orgs SET quota_num_users = 10, quota_managed_data_bytes = 5368709120; -- review: does this fit with current limits

ALTER TABLE projects ADD COLUMN next_usage_reporting_time TIMESTAMP DEFAULT '0001-01-01 00:00:00';
