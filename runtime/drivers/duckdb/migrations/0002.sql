ALTER TABLE rill.catalog ADD COLUMN refreshed_on TIMESTAMPTZ;
ALTER TABLE rill.catalog ADD COLUMN definition VARBINARY;
ALTER TABLE rill.catalog ADD COLUMN path TEXT;
CREATE UNIQUE INDEX lower_name_unique_idx ON rill.catalog (lower(name));

UPDATE rill.catalog SET refreshed_on = created_on;