-- OAuth state for connectors that authorize users against an external OAuth server (e.g. driver "mcp" with auth "oauth").

-- An OAuth client Rill uses with a connector's authorization server: discovered endpoints plus the registered (or pre-registered) client,
-- shared by the users of a project's connector.
-- A client is bound to a hash of the connector's OAuth configuration (as computed by the runtime),
-- so changing the configuration (URL, pre-registered client, scopes) leads to a new client and never reuses authorizations granted for the old one.
-- Clients are immutable.
CREATE TABLE connector_oauth_clients (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    connector TEXT NOT NULL,
    config_hash TEXT NOT NULL,
    -- JSON of the client (see runtime/pkg/oauthclient.Client), encrypted when a key ID is set
    client TEXT NOT NULL,
    client_encryption_key_id TEXT NOT NULL DEFAULT '',
    created_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX connector_oauth_clients_project_id_connector_config_hash_idx ON connector_oauth_clients (project_id, connector, config_hash);

-- A user's OAuth token, obtained through a client. Tokens are never shared between users.
-- A token is only used (and refreshed) through the client that obtained it.
CREATE TABLE connector_oauth_tokens (
    id UUID NOT NULL DEFAULT uuid_generate_v4() PRIMARY KEY,
    oauth_client_id UUID NOT NULL REFERENCES connector_oauth_clients (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- JSON of the token (access token, refresh token, expiry), encrypted when a key ID is set
    token TEXT NOT NULL,
    token_encryption_key_id TEXT NOT NULL DEFAULT '',
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX connector_oauth_tokens_oauth_client_id_user_id_idx ON connector_oauth_tokens (oauth_client_id, user_id);
CREATE INDEX connector_oauth_tokens_user_id_idx ON connector_oauth_tokens (user_id);
