ALTER TABLE projects
ADD COLUMN production_branch TEXT,
ADD COLUMN github_url TEXT,
ADD COLUMN github_installation_id BIGINT;

CREATE UNIQUE INDEX projects_github_url_idx ON projects (lower(github_url));
