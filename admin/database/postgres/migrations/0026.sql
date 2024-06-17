ALTER TABLE projects ADD COLUMN created_by_user_id UUID REFERENCES users (id) ON DELETE SET NULL;
