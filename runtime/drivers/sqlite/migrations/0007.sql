ALTER TABLE instances ADD COLUMN olap_connector TEXT NOT NULL DEFAULT '';
UPDATE instances SET olap_connector = olap_driver;
ALTER TABLE instances DROP COLUMN olap_driver; 

ALTER TABLE instances ADD COLUMN repo_connector TEXT NOT NULL DEFAULT '';
UPDATE instances SET repo_connector = repo_driver;
ALTER TABLE instances DROP COLUMN repo_driver;
