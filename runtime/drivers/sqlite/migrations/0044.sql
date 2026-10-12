CREATE TABLE IF NOT EXISTS ai_memories (
    id TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    agent TEXT NOT NULL,
    content TEXT NOT NULL,
    source_session_id TEXT NOT NULL,
    created_on TIMESTAMP NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS ai_memories_instance_id_owner_id_idx ON ai_memories (instance_id, owner_id);

CREATE TABLE IF NOT EXISTS ai_memory_settings (
    instance_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    enabled BOOLEAN NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    PRIMARY KEY (instance_id, owner_id)
);
