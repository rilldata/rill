CREATE TABLE ai_feedback (
    id TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    target_message_id TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL,
    sentiment TEXT NOT NULL DEFAULT '',
    categories TEXT NOT NULL DEFAULT '[]',
    comment TEXT NOT NULL DEFAULT '',
    predicted_attribution TEXT NOT NULL DEFAULT '',
    suggested_action TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open',
    owner_id TEXT NOT NULL DEFAULT '',
    owner_email TEXT NOT NULL DEFAULT '',
    resolved_by TEXT NOT NULL DEFAULT '',
    resolved_on TIMESTAMP,
    created_on TIMESTAMP NOT NULL,
    updated_on TIMESTAMP NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (session_id) REFERENCES ai_sessions(id) ON DELETE CASCADE
);

CREATE INDEX ai_feedback_instance_status_created_idx ON ai_feedback (instance_id, status, created_on);

CREATE INDEX ai_feedback_session_id_idx ON ai_feedback (session_id);
