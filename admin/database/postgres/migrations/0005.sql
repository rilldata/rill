ALTER TABLE orgs ADD COLUMN quota_projects INTEGER NOT NULL DEFAULT -1;
ALTER TABLE orgs ADD COLUMN quota_deployments INTEGER NOT NULL DEFAULT -1;
ALTER TABLE orgs ADD COLUMN quota_slots_total INTEGER NOT NULL DEFAULT -1;
ALTER TABLE orgs ADD COLUMN quota_slots_per_deployment INTEGER NOT NULL DEFAULT -1;
ALTER TABLE orgs ADD COLUMN quota_outstanding_invites INTEGER NOT NULL DEFAULT -1;
UPDATE orgs SET quota_projects = 5, quota_deployments = 10, quota_slots_total = 20, quota_slots_per_deployment = 5, quota_outstanding_invites = 200;

ALTER TABLE users ADD COLUMN quota_singleuser_orgs INTEGER NOT NULL DEFAULT -1;
UPDATE users SET quota_singleuser_orgs = 3;