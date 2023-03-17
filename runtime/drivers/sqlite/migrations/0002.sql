CREATE TABLE catalog (
    instance_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type INTEGER NOT NULL,
    object BLOB NOT NULL,
    path TEXT NOT NULL,
    created_on TIMESTAMP NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    refreshed_on TIMESTAMP NOT NULL,
    PRIMARY KEY (instance_id, name)
);

CREATE UNIQUE INDEX lower_name_unique_idx ON catalog (instance_id, lower(name));
