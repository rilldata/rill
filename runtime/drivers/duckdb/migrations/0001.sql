CREATE TABLE rill.catalog (
	name TEXT NOT NULL,
	type INTEGER NOT NULL,
	object BLOB NOT NULL,
	path TEXT NOT NULL,
	created_on TIMESTAMPTZ NOT NULL,
	updated_on TIMESTAMPTZ NOT NULL,
	refreshed_on TIMESTAMPTZ NOT NULL,
	PRIMARY KEY (name)
);
