CREATE TABLE rill.catalogv2 (
	kind TEXT NOT NULL,
	name TEXT NOT NULL,
	data BLOB NOT NULL,
	created_on TIMESTAMPTZ NOT NULL,
	updated_on TIMESTAMPTZ NOT NULL
);

CREATE TABLE rill.controller_version (
	version INTEGER NOT NULL,
);

INSERT INTO rill.controller_version (version) VALUES (0);
