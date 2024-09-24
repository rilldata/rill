PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS instance_health (
    instance_id TEXT PRIMARY KEY,
    health BLOB NOT NULL,
    created_on TIMESTAMP NOT NULL,
    FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE
);