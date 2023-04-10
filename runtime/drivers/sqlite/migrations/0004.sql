ALTER TABLE instances ADD COLUMN ingestion_limit_bytes integer default 0 NOT NULL;
ALTER TABLE catalog ADD COLUMN bytes_ingested INTEGER default 0 NOT NULL;