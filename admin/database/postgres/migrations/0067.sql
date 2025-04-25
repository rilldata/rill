CREATE TABLE IF NOT EXISTS managed_github_repo_meta (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    org_id UUID NOT NULL REFERENCES orgs (id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects (id) ON DELETE SET NULL,
    html_url TEXT NOT NULL UNIQUE,
    created_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    created_on TIMESTAMP DEFAULT NOW(),
    updated_on TIMESTAMP DEFAULT NOW()
);

CREATE INDEX managed_github_repo_meta_project_id_idx ON managed_github_repo_meta (project_id);
