DROP INDEX bookmarks_search_idx;
CREATE INDEX bookmarks_search_idx ON bookmarks (project_id, resource_kind, lower(resource_name));
