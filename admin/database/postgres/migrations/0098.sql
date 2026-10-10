-- OAuth state for external MCP servers (connectors with driver "mcp" and auth "oauth").

-- The OAuth client Rill uses with a server's authorization server: discovered endpoints plus the registered (or pre-registered) client.
-- One per project and connector, shared by all users of the connector.
CREATE TABLE mcp_oauth_clients (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    connector TEXT NOT NULL,
    -- JSON of the client (see runtime/drivers/mcp.OAuthClient), encrypted when a key ID is set
    client TEXT NOT NULL,
    client_encryption_key_id TEXT NOT NULL DEFAULT '',
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX mcp_oauth_clients_project_id_connector_idx ON mcp_oauth_clients (project_id, connector);

-- A user's OAuth token for a server. Tokens are never shared between users.
CREATE TABLE mcp_oauth_tokens (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    connector TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- JSON of the token (access token, refresh token, expiry), encrypted when a key ID is set
    token TEXT NOT NULL,
    token_encryption_key_id TEXT NOT NULL DEFAULT '',
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX mcp_oauth_tokens_project_id_connector_user_id_idx ON mcp_oauth_tokens (project_id, connector, user_id);
