-- Lets an org refuse new project domain access rules (projects_autoinvite_domains); existing rules keep working.
ALTER TABLE orgs ADD COLUMN disable_project_domain_access BOOLEAN NOT NULL DEFAULT false;
