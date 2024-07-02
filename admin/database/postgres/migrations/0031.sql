ALTER TABLE orgs ADD COLUMN billing_customer_id TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX orgs_billing_customer_id_idx ON orgs (billing_customer_id) WHERE billing_customer_id <> '';

ALTER TABLE orgs ADD COLUMN quota_storage_limit_bytes_per_deployment BIGINT NOT NULL DEFAULT -1;
UPDATE orgs SET quota_storage_limit_bytes_per_deployment = 5368709120;

CREATE TABLE billing_reporting_time( usage_reported_on TIMESTAMPTZ );

INSERT INTO billing_reporting_time(usage_reported_on) VALUES (NULL);
