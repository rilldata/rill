PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS instance_health (
    instance_id TEXT PRIMARY KEY,
    health_json BLOB NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    FOREIGN KEY (instance_id) REFERENCES instances(id) ON DELETE CASCADE
);