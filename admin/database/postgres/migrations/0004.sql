ALTER TABLE projects
ADD COLUMN production_branch TEXT,
ADD COLUMN github_url TEXT,
ADD COLUMN github_installation_id BIGINT;

CREATE UNIQUE INDEX projects_github_url_idx ON projects (lower(github_url));

CREATE TABLE users_github_installations (
	user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	installation_id BIGINT NOT NULL,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
	PRIMARY KEY (user_id, installation_id)
);
