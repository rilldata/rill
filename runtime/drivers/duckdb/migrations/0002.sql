ALTER TABLE rill.catalog ADD COLUMN refreshed_on TIMESTAMPTZ;

UPDATE rill.catalog SET refreshed_on = created_on;