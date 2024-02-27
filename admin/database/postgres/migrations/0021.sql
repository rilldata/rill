ALTER TABLE bookmarks ADD COLUMN default boolean DEFAULT false;
ALTER TABLE bookmarks ADD COLUMN shared boolean DEFAULT false;
ALTER TABLE bookmarks ADD COLUMN description TEXT DEFAULT '';

CREATE INDEX bookmarks_project_id_dashboard_name ON bookmarks (project_id, dashboard_name);
