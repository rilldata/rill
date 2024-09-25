CREATE TABLE report_tokens (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    report_name TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    magic_auth_token_id UUID NOT NULL REFERENCES magic_auth_tokens (id) ON DELETE CASCADE
);