CREATE TABLE IF NOT EXISTS conversations (
    instance_id TEXT NOT NULL,
    conversation_id TEXT NOT NULL,
    title TEXT,
    created_on TIMESTAMP NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    PRIMARY KEY (instance_id, conversation_id)
);

CREATE TABLE IF NOT EXISTS messages (
    instance_id TEXT NOT NULL,
    conversation_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    role TEXT NOT NULL,
    content_json TEXT NOT NULL,
    created_on TIMESTAMP NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    PRIMARY KEY (instance_id, conversation_id, message_id),
    FOREIGN KEY (instance_id, conversation_id) REFERENCES conversations(instance_id, conversation_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_instance_id ON messages (instance_id);
CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages (conversation_id);
