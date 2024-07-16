ALTER TABLE orgs ADD COLUMN payment_customer_id TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX orgs_payment_customer_id_idx ON orgs (payment_customer_id) WHERE payment_customer_id <> '';
