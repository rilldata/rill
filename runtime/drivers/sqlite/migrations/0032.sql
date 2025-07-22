CREATE TABLE IF NOT EXISTS conversations (
    instance_id TEXT NOT NULL,
    conversation_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    title TEXT,
    created_on TIMESTAMP NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    PRIMARY KEY (instance_id, conversation_id)
);

CREATE TABLE IF NOT EXISTS messages (
    instance_id TEXT NOT NULL,
    conversation_id TEXT NOT NULL,
    seq_num INTEGER NOT NULL,
    message_id TEXT NOT NULL UNIQUE,
    role TEXT NOT NULL,
    content_json TEXT NOT NULL,
    created_on TIMESTAMP NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    PRIMARY KEY (instance_id, conversation_id, seq_num),
    FOREIGN KEY (instance_id, conversation_id) REFERENCES conversations(instance_id, conversation_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_conversations_instance_owner ON conversations (instance_id, owner_id);
