CREATE TABLE billing_errors (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    org_id UUID NOT NULL,
    type INTEGER NOT NULL,
    msg TEXT NOT NULL,
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (org_id) REFERENCES orgs (id) ON DELETE CASCADE,
    CONSTRAINT billing_errors_org_id_error_type_unique UNIQUE (org_id, type)
);

CREATE TABLE billing_warnings (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    org_id UUID NOT NULL,
    type INTEGER NOT NULL,
    msg TEXT NOT NULL,
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (org_id) REFERENCES orgs (id) ON DELETE CASCADE,
    CONSTRAINT billing_warnings_org_id_warning_type_unique UNIQUE (org_id, type)
);