ALTER TABLE orgs ADD COLUMN payment_customer_id TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX orgs_payment_customer_id_idx ON orgs (payment_customer_id) WHERE payment_customer_id <> '';

ALTER TABLE billing_reporting_time RENAME TO billing_worker_time;

ALTER TABLE billing_worker_time ADD COLUMN repaired_on TIMESTAMPTZ DEFAULT NULL;
