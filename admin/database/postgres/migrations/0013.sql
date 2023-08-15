ALTER TABLE service ADD COLUMN active_on TIMESTAMPTZ DEFAULT now() NOT NULL;

UPDATE service SET active_on = updated_on;
