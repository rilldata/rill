-- OAuth state for external MCP servers (connectors with driver "mcp" and auth "oauth").

-- An OAuth client Rill uses with a server's authorization server: discovered endpoints plus the registered (or pre-registered) client,
-- shared by the users of a project's connector.
-- Rows are keyed by server URL, so a connector (or a branch's copy of it) that is pointed at another server gets its own client
-- and tokens granted for the previous server are never refreshed or revoked against the new one.
-- A client is never modified once stored, because tokens and pending authorizations are bound to it:
-- a change to the connector's OAuth configuration (scopes, pre-registered client) registers a new client next to the old one,
-- and clients that nothing is bound to anymore are deleted.
CREATE TABLE mcp_oauth_clients (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    connector TEXT NOT NULL,
    server_url TEXT NOT NULL,
    -- JSON of the client (see runtime/drivers/mcp.OAuthClient), encrypted when a key ID is set
    client TEXT NOT NULL,
    client_encryption_key_id TEXT NOT NULL DEFAULT '',
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX mcp_oauth_clients_project_id_connector_server_url_idx ON mcp_oauth_clients (project_id, connector, server_url);

-- A user's OAuth token for a server. Tokens are never shared between users.
-- Each token is bound to the client that obtained it; it is only used (and refreshed or revoked) through that client.
CREATE TABLE mcp_oauth_tokens (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    connector TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    oauth_client_id UUID NOT NULL REFERENCES mcp_oauth_clients (id) ON DELETE CASCADE,
    -- JSON of the token (access token, refresh token, expiry), encrypted when a key ID is set
    token TEXT NOT NULL,
    token_encryption_key_id TEXT NOT NULL DEFAULT '',
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX mcp_oauth_tokens_project_id_connector_user_id_idx ON mcp_oauth_tokens (project_id, connector, user_id);
CREATE INDEX mcp_oauth_tokens_oauth_client_id_idx ON mcp_oauth_tokens (oauth_client_id);
