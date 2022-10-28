ALTER TABLE rill.catalog ADD COLUMN refreshed_on TIMESTAMPTZ;
CREATE UNIQUE INDEX lower_name_unique_idx ON rill.catalog (lower(name));

UPDATE rill.catalog SET refreshed_on = created_on;