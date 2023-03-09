ALTER TABLE projects
ADD COLUMN public BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN production_branch TEXT,
ADD COLUMN github_url TEXT,
ADD COLUMN github_installation_id BIGINT;

CREATE UNIQUE INDEX projects_github_url_idx ON projects (lower(github_url));

CREATE TABLE users_github_installations (
	id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	installation_id BIGINT NOT NULL,
	created_on TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE UNIQUE INDEX users_github_installations_user_id_installation_id_idx ON users_github_installations (user_id, installation_id);
