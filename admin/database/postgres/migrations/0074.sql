ALTER TABLE projects ADD COLUMN dev_slots INTEGER NOT NULL;
ALTER TABLE projects ADD COLUMN dev_ttl_seconds BIGINT NOT NULL; 
ALTER TABLE deployments ADD COLUMN environment TEXT NOT NULL;
ALTER TABLE deployments ADD COLUMN owner_user_id UUID REFERENCES users(id) ON DELETE SET NULL;

UPDATE projects SET dev_slots = 8;
UPDATE projects SET dev_ttl_seconds = 21600; -- 6 hours
UPDATE deployments SET environment = 'prod';
