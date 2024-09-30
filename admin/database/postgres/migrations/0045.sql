CREATE TABLE billing_issues (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    org_id UUID NOT NULL,
    type INTEGER NOT NULL,
    level INTEGER NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    overdue_processed BOOLEAN NOT NULL DEFAULT FALSE,
    event_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (org_id) REFERENCES orgs (id) ON DELETE CASCADE,
    CONSTRAINT billing_issues_org_id_type_unique UNIQUE (org_id, type)
);