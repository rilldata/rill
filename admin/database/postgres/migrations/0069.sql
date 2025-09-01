CREATE TABLE managed_git_repos (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    org_id UUID REFERENCES orgs (id) ON DELETE SET NULL,
    remote TEXT NOT NULL UNIQUE,
    owner_id UUID NOT NULL,
    created_on TIMESTAMP DEFAULT NOW(),
    updated_on TIMESTAMP DEFAULT NOW()
);

ALTER TABLE PROJECTS ADD column managed_git_repo_id UUID REFERENCES managed_git_repos (id) ON DELETE SET NULL;
ALTER TABLE PROJECTS ADD column github_repo_id BIGINT;