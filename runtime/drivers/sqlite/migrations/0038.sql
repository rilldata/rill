ALTER TABLE ai_sessions ADD COLUMN shared_until_message_id TEXT NOT NULL DEFAULT '';
ALTER TABLE ai_sessions ADD COLUMN forked_from_session_id TEXT NOT NULL DEFAULT '';