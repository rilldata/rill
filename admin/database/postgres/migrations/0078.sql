CREATE TABLE alert_tokens (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    alert_name TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    magic_auth_token_id UUID NOT NULL REFERENCES magic_auth_tokens (id) ON DELETE CASCADE
);

CREATE INDEX alert_tokens_alert_name_idx ON alert_tokens (alert_name);
CREATE INDEX alert_tokens_magic_auth_token_id_idx ON alert_tokens (magic_auth_token_id);
