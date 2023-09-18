ALTER TABLE instances ADD COLUMN watch_repo BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE instances ADD COLUMN stage_changes BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE instances ADD COLUMN model_default_materialize BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE instances ADD COLUMN model_materialize_delay_seconds INTEGER NOT NULL DEFAULT 0;
