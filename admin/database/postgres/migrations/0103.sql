-- User-defined labels for organizing projects (e.g. "prod", "sandbox").
-- Unlike annotations, tags are managed by org users and never propagate to deployments.
-- Every query on tags is scoped to an org, so they don't need an index.
ALTER TABLE projects ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}'::TEXT[];
