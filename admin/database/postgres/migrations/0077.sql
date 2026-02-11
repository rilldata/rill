-- Add column "directory_name" to the "projects" table.
ALTER TABLE projects ADD COLUMN directory_name TEXT NOT NULL DEFAULT '';