ALTER TABLE projects
ADD COLUMN production_olap_driver TEXT NOT NULL DEFAULT 'duckdb',
ADD COLUMN production_olap_dsn TEXT NOT NULL DEFAULT '';
