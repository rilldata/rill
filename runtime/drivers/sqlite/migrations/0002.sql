CREATE TABLE catalog (
    instance_id TEXT,
	name TEXT,
	type TEXT NOT NULL,
	sql TEXT NOT NULL,
    schema BLOB,
    managed BOOLEAN NOT NULL,
	created_on TIMESTAMP NOT NULL,
	updated_on TIMESTAMP NOT NULL,
    refreshed_on TIMESTAMP NOT NULL,
    PRIMARY KEY (instance_id, name)
);

CREATE UNIQUE INDEX lower_name_unique_idx ON catalog (instance_id, lower(name));
