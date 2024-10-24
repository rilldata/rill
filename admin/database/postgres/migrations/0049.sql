ALTER TABLE magic_auth_tokens ADD COLUMN internal BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE report_tokens (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    report_name TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    magic_auth_token_id UUID NOT NULL REFERENCES magic_auth_tokens (id) ON DELETE CASCADE
);

CREATE INDEX report_tokens_report_name_idx ON report_tokens (report_name);
CREATE INDEX report_tokens_magic_auth_token_id_idx ON report_tokens (magic_auth_token_id);
