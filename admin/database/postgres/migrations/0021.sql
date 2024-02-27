ALTER TABLE bookmarks RENAME COLUMN dashboard_name TO resource_name;
ALTER TABLE bookmarks ADD COLUMN resource_kind TEXT DEFAULT 'MetricsView';
ALTER TABLE bookmarks ADD COLUMN "default" boolean DEFAULT false;
ALTER TABLE bookmarks ADD COLUMN shared boolean DEFAULT false;
ALTER TABLE bookmarks ADD COLUMN description TEXT DEFAULT '';

CREATE INDEX bookmarks_search_idx ON bookmarks (project_id, user_id, resource_name, resource_kind, "default", shared);
