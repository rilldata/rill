ALTER TABLE rill.catalog ADD COLUMN embedded BOOL;
UPDATE rill.catalog SET embedded = FALSE WHERE embedded IS NULL;
