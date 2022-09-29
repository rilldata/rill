CREATE TABLE catalog (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL
);

CREATE UNIQUE INDEX catalog_name_idx ON catalog (name);
