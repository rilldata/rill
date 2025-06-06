-- Rename projects.github_url to projects.git_remote
ALTER TABLE projects RENAME COLUMN github_url TO git_remote;

-- Repeat migration 0070.sql to compensate for bug.
UPDATE projects
SET git_remote = concat(git_remote, '.git')
WHERE git_remote IS NOT NULL
  AND git_remote NOT LIKE '%.git';
