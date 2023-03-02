ALTER TABLE projects
ADD COLUMN git_url TEXT UNIQUE,
ADD COLUMN git_full_name TEXT UNIQUE,
ADD COLUMN github_app_install_id BIGINT,  
ADD COLUMN production_branch TEXT;