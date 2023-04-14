DROP INDEX projects_github_url_idx;
CREATE INDEX projects_github_url_idx ON projects lower(github_url, org_id);