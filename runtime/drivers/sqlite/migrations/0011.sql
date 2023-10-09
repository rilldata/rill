DROP TABLE IF EXISTS catalogv2;
DROP TABLE IF EXISTS controller_version;

CREATE TABLE catalogv2 (
    instance_id TEXT NOT NULL,
	kind TEXT NOT NULL,
	name TEXT NOT NULL,
	data BLOB NOT NULL,
	created_on TIMESTAMPTZ NOT NULL,
	updated_on TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX catalogv2_instance_id_name_idx ON catalog (instance_id, kind, lower(name));

CREATE TABLE controller_version (
    instance_id TEXT NOT NULL,
	version INTEGER NOT NULL
);

CREATE UNIQUE INDEX controller_version_instance_id_idx ON catalog (instance_id);

INSERT INTO controller_version (version) VALUES (0);
