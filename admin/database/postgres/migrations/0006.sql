ALTER TABLE users ADD COLUMN superuser BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE orgs_autoinvite_domains (
    id uuid not null primary key default uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES orgs (id) ON DELETE CASCADE,
    org_role_id UUID NOT NULL REFERENCES org_roles (id) ON DELETE CASCADE,
    domain TEXT NOT NULL,
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX orgs_autoinvite_domains_domain_idx ON orgs_autoinvite_domains (lower(domain));
CREATE UNIQUE INDEX orgs_autoinvite_domains_org_id_domain_idx ON orgs_autoinvite_domains (org_id, lower(domain));