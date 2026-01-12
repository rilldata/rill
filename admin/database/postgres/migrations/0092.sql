CREATE TABLE slack_workspaces (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	team_id TEXT NOT NULL UNIQUE,
	team_name TEXT,
	customer_id TEXT, -- Optional: link to org_id or external customer identifier
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE INDEX slack_workspaces_team_id_idx ON slack_workspaces (team_id);
CREATE INDEX slack_workspaces_customer_id_idx ON slack_workspaces (customer_id) WHERE customer_id IS NOT NULL;

CREATE TABLE slack_user_tokens (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES slack_workspaces (id) ON DELETE CASCADE,
	user_id TEXT NOT NULL, -- Slack user ID
	token_encrypted BYTEA NOT NULL, -- Encrypted Rill PAT
	token_encryption_key_id TEXT, -- ID of encryption key used (for key rotation)
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	updated_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	UNIQUE(workspace_id, user_id)
);

CREATE INDEX slack_user_tokens_workspace_user_idx ON slack_user_tokens (workspace_id, user_id);

CREATE TABLE slack_conversations (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	workspace_id UUID NOT NULL REFERENCES slack_workspaces (id) ON DELETE CASCADE,
	channel_id TEXT NOT NULL,
	thread_ts TEXT, -- NULL for non-threaded conversations
	rill_conversation_id TEXT NOT NULL,
	user_id TEXT NOT NULL, -- Slack user ID
	last_activity TIMESTAMPTZ DEFAULT now() NOT NULL,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	UNIQUE(workspace_id, channel_id, thread_ts)
);

CREATE INDEX slack_conversations_workspace_channel_idx ON slack_conversations (workspace_id, channel_id);
CREATE INDEX slack_conversations_workspace_channel_thread_idx ON slack_conversations (workspace_id, channel_id, thread_ts);
