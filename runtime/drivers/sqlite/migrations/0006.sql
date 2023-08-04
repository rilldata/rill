ALTER TABLE instances ADD COLUMN connectors TEXT NOT NULL DEFAULT '';
UPDATE instances SET connectors = FORMAT('[{"type":"%s","name":"repo","configs":{"dsn":"%s"}},{"type":"%s","name":"olap","configs":{"dsn":"%s"}}]', repo_driver, repo_dsn, olap_driver, olap_dsn);
UPDATE instances SET olap_driver = 'olap';
UPDATE instances SET repo_driver = 'repo';
ALTER TABLE instances DROP COLUMN olap_dsn; 
ALTER TABLE instances DROP COLUMN repo_dsn;