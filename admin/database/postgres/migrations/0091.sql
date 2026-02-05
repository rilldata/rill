ALTER TABLE magic_auth_tokens ADD COLUMN metrics_view_filter_jsons JSONB NOT NULL DEFAULT '{}'::JSONB;
-- migrate data from filter_json column to metrics_view_filter_jsons column; since it won't have any metrics view names, we can set the key to '*' to indicate all metrics views
UPDATE magic_auth_tokens SET metrics_view_filter_jsons = jsonb_build_object('*', filter_json) where filter_json != '{}' AND filter_json != '';
-- drop the old filter_json column
ALTER TABLE magic_auth_tokens DROP COLUMN filter_json;
