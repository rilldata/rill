ALTER TABLE bookmarks ADD COLUMN is_global boolean DEFAULT false;

CREATE INDEX bookmarks_project_id_dashboard_name ON bookmarks (project_id, dashboard_name);
