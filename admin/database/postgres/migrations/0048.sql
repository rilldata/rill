ALTER TABLE users ADD COLUMN quota_trial_orgs INTEGER NOT NULL DEFAULT -1;
ALTER TABLE users ADD COLUMN current_trial_orgs_count INTEGER NOT NULL DEFAULT 0;
UPDATE users SET quota_trial_orgs = 2;
UPDATE users SET quota_singleuser_orgs = 100;

ALTER TABLE orgs ADD COLUMN created_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL;
