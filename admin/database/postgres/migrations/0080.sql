-- Rename reports_token table to notification_tokens
ALTER TABLE report_tokens RENAME TO notification_tokens;

-- Add new columns for general resource support with temporary defaults for migration
ALTER TABLE notification_tokens ADD COLUMN resource_kind TEXT DEFAULT '' NOT NULL;
ALTER TABLE notification_tokens ADD COLUMN resource_name TEXT DEFAULT '' NOT NULL;

-- Migrate existing data: copy report_name to resource_name
UPDATE notification_tokens SET resource_kind = 'rill.runtime.v1.Report', resource_name = report_name;

-- Drop the old report_name column
ALTER TABLE notification_tokens DROP COLUMN report_name;

-- Remove the temporary defaults since these should be required fields
ALTER TABLE notification_tokens ALTER COLUMN resource_kind DROP DEFAULT;
ALTER TABLE notification_tokens ALTER COLUMN resource_name DROP DEFAULT;
