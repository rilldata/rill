ALTER TABLE rill.catalog ADD COLUMN refreshed_on TIMESTAMPTZ;
UPDATE rill.catalog SET refreshed_on = created_on;
CREATE UNIQUE INDEX lower_name_unique_idx ON rill.catalog (lower(name));
