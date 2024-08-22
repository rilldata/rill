CREATE TABLE billing_errors (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    org_id UUID NOT NULL,
    type INTEGER NOT NULL,
    msg TEXT NOT NULL,
    triggers_river_job_id BIGINT NOT NULL,
    event_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (org_id) REFERENCES orgs (id) ON DELETE CASCADE,
    CONSTRAINT billing_errors_org_id_type_unique UNIQUE (org_id, type)
);

CREATE TABLE billing_warnings (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    org_id UUID NOT NULL,
    type INTEGER NOT NULL,
    msg TEXT NOT NULL,
    triggers_river_job_id BIGINT NOT NULL,
    event_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (org_id) REFERENCES orgs (id) ON DELETE CASCADE,
    CONSTRAINT billing_warnings_org_id_type_unique UNIQUE (org_id, type)
);

CREATE TABLE webhook_event_watermarks (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    org_id UUID NOT NULL,
    type TEXT NOT NULL,
    last_occurrence TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (org_id) REFERENCES orgs (id) ON DELETE CASCADE,
    CONSTRAINT webhook_event_watermark_org_id_type_unique UNIQUE (org_id, type)
);
