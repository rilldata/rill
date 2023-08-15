ALTER TABLE instances ADD COLUMN connectors TEXT NOT NULL DEFAULT '';
ALTER TABLE instances ADD COLUMN project_connectors TEXT NOT NULL DEFAULT '';
UPDATE
	instances
SET
	connectors = format('[%s, %s]',
	json_replace('{"type":"%s","name":"repo","config":{"dsn":"%s"}}',
	'$.type',
	repo_driver,
	'$.config.dsn',
	repo_dsn),
	json_replace('{"type":"%s","name":"olap","config":{"dsn":"%s"}}',
	'$.type',
	olap_driver,
	'$.config.dsn',
	olap_dsn));
UPDATE instances SET olap_driver = 'olap';
UPDATE instances SET repo_driver = 'repo';
ALTER TABLE instances DROP COLUMN olap_dsn; 
ALTER TABLE instances DROP COLUMN repo_dsn;
