DROP INDEX bookmarks_search_idx;
CREATE INDEX bookmarks_search_idx ON bookmarks (project_id, user_id, lower(resource_name), resource_kind, "default", shared);