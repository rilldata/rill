CREATE TABLE managed_git_repo (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    remote TEXT NOT NULL UNIQUE,
    created_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL,
    created_on TIMESTAMP DEFAULT NOW(),
    updated_on TIMESTAMP DEFAULT NOW()
);

CREATE INDEX managed_git_repo_project_id_idx ON managed_git_repo (project_id);

ALTER TABLE PROJECTS ADD column managed_git_repo_id UUID REFERENCES managed_git_repo (id) ON DELETE RESTRICT;
ALTER TABLE PROJECTS ADD column github_repository_id BIGINT;