ALTER TABLE magic_auth_tokens ADD COLUMN resource_type TEXT NOT NULL DEFAULT 'rill.runtime.v1.Explore';
ALTER TABLE magic_auth_tokens RENAME COLUMN metrics_view TO resource_name;
ALTER TABLE magic_auth_tokens RENAME COLUMN metrics_view_filter_json TO filter_json;
ALTER TABLE magic_auth_tokens RENAME COLUMN metrics_view_fields TO fields;

UPDATE bookmarks SET resource_kind = 'rill.runtime.v1.Explore' WHERE resource_kind = 'rill.runtime.v1.MetricsView';
