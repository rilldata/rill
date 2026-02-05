DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS messages;

CREATE TABLE IF NOT EXISTS ai_sessions (
    id TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    title TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    created_on TIMESTAMP NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS ai_sessions_instance_id_owner_id_idx ON ai_sessions (instance_id, owner_id);

CREATE TABLE IF NOT EXISTS ai_messages (
    id TEXT NOT NULL,
    parent_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    created_on TIMESTAMP NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    "index" INTEGER NOT NULL,
    "role" TEXT NOT NULL,
    "type" TEXT NOT NULL,
    tool TEXT NOT NULL,
    content_type TEXT NOT NULL,
    content TEXT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (session_id) REFERENCES ai_sessions(id) ON DELETE CASCADE
);
