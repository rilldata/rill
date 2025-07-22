-- Deprecate the use of prod_olap_driver and prod_olap_dsn in the projects table
ALTER TABLE projects DROP COLUMN IF EXISTS prod_olap_driver;
ALTER TABLE projects DROP COLUMN IF EXISTS prod_olap_dsn;
