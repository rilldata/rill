ALTER TABLE users ADD COLUMN preference_timezone TEXT NOT NULL DEFAULT '';

CREATE TABLE dashboard_bookmarks (
  id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
  display_name TEXT NOT NULL,
  data BYTEA NOT NULL,
  dashboard_name TEXT NOT NULL,
  project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
  updated_on TIMESTAMPTZ DEFAULT now() NOT NULL
);
