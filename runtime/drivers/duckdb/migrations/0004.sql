-- DuckDB cannot add columns to a table with indexes. So we drop
-- lower_name_unique_idx, add the columns, then create lower_name_unique_idx again.

DROP INDEX IF EXISTS rill.lower_name_unique_idx;
ALTER TABLE rill.catalog ALTER bytes_ingested TYPE BIGINT;
CREATE UNIQUE INDEX lower_name_unique_idx ON rill.catalog (lower(name));
