CREATE TABLE instances (
	id TEXT PRIMARY KEY,
	olap_driver TEXT NOT NULL,
	olap_dsn TEXT NOT NULL,
	repo_driver TEXT NOT NULL,
	repo_dsn TEXT NOT NULL,	
	embed_catalog BOOLEAN NOT NULL,
	created_on TIMESTAMP NOT NULL,
	updated_on TIMESTAMP NOT NULL,
	variables TEXT,
	project_variables TEXT
);
