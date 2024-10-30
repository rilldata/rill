CREATE TABLE model_splits (
    instance_id TEXT NOT NULL,
    model_id TEXT NOT NULL,
    key TEXT NOT NULL,
    data_json BLOB NOT NULL,
    idx INTEGER NOT NULL,
    watermark TIMESTAMP,
    executed_on TIMESTAMP,
    error TEXT,
    elapsed_ms INTEGER,
    PRIMARY KEY (instance_id, model_id, key)
);
