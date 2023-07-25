ALTER TABLE rill.catalog ADD COLUMN bytes_ingested INTEGER;
UPDATE rill.catalog SET bytes_ingested = 0 WHERE bytes_ingested IS NULL;
