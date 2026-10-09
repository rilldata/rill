CREATE TABLE IF NOT EXISTS ai_memories (
    id TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    category TEXT NOT NULL,
    content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    source TEXT NOT NULL DEFAULT 'extracted',
    source_session_id TEXT NOT NULL DEFAULT '',
    source_message_id TEXT NOT NULL DEFAULT '',
    created_on TIMESTAMP NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS ai_memories_instance_id_owner_id_status_idx ON ai_memories (instance_id, owner_id, status);

CREATE TABLE IF NOT EXISTS ai_memory_settings (
    instance_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    paused BOOLEAN NOT NULL DEFAULT false,
    updated_on TIMESTAMP NOT NULL,
    PRIMARY KEY (instance_id, owner_id)
);

ALTER TABLE ai_sessions ADD COLUMN memory_disabled BOOLEAN NOT NULL DEFAULT false;
