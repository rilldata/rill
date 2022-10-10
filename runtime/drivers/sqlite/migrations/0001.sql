CREATE TABLE instances (
	id TEXT PRIMARY KEY,
	driver TEXT NOT NULL,
	dsn TEXT NOT NULL,
	object_prefix TEXT NOT NULL,
	exposed BOOLEAN NOT NULL,
	embed_catalog BOOLEAN NOT NULL,
	created_on TIMESTAMP NOT NULL,
	updated_on TIMESTAMP NOT NULL
);

CREATE TABLE repos (
	id TEXT PRIMARY KEY,
	driver TEXT NOT NULL,
	dsn TEXT NOT NULL,
	created_on TIMESTAMP NOT NULL,
	updated_on TIMESTAMP NOT NULL
);
