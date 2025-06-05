-- Rename projects.github_url to projects.git_remote
ALTER TABLE projects  RENAME COLUMN github_url TO git_remote;
