ALTER TABLE orgs ADD COLUMN quota_projects INTEGER DEFAULT 5;
ALTER TABLE orgs ADD COLUMN quota_deployments INTEGER DEFAULT 10;
ALTER TABLE orgs ADD COLUMN quota_slots_total INTEGER DEFAULT 20;
ALTER TABLE orgs ADD COLUMN quota_slots_per_deployment INTEGER DEFAULT 5;
ALTER TABLE orgs ADD COLUMN quota_outstanding_invitations INTEGER DEFAULT 200;

ALTER TABLE users ADD COLUMN quota_singleuser_orgs INTEGER DEFAULT 3;