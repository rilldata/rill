CREATE TABLE rill.catalog (
	name TEXT NOT NULL,
	type INTEGER NOT NULL,
	object BLOB NOT NULL,
	path TEXT NOT NULL,
	embedded BOOL,
	created_on TIMESTAMPTZ NOT NULL,
	updated_on TIMESTAMPTZ NOT NULL,
	refreshed_on TIMESTAMPTZ NOT NULL,
	PRIMARY KEY (name)
);

CREATE UNIQUE INDEX lower_name_unique_idx ON rill.catalog (lower(name));
