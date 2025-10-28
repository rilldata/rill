ALTER TABLE deployments
DROP CONSTRAINT deployments_project_id_fkey;

ALTER TABLE deployments
ALTER COLUMN project_id DROP NOT NULL,
ADD CONSTRAINT deployments_project_id_fkey FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;
