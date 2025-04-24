CREATE OR REPLACE TABLE managed_github_repo_meta (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs (id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects (id) ON DELETE SET NULL,
    created_by UUID NOT NULL REFERENCES users (id) ON DELETE SET NULL,
    html_url TEXT NOT NULL DISTINCT,
    created_on TIMESTAMP DEFAULT NOW(),
    updated_on TIMESTAMP DEFAULT NOW()
)