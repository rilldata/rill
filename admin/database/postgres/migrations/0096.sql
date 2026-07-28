-- Track the Git-sync state of each virtual file.
-- Virtual files act as a staging buffer: writes land here instantly and are later
-- promoted into the project's Git repo under user_files/<segment>/<kind>/<name>.yaml.
-- sync_state: 'dirty' (staged, not yet in Git) or 'synced' (content confirmed on the served branch;
-- the row is kept as a path index for later edits). 'pending_pr' is reserved for a future pull request flow.
ALTER TABLE virtual_files ADD COLUMN sync_state TEXT NOT NULL DEFAULT 'dirty';

-- Snapshot of the file's content when it last transitioned from synced to dirty, i.e. the version the
-- staged change was based on. NULL for rows that were never synced. The sync job compares it to the
-- Git copy to detect (and report) when a sync overwrites edits made directly in Git while the row was dirty.
ALTER TABLE virtual_files ADD COLUMN base_data BYTEA;

-- Cheap lookups of a project's staged (unsynced) files for the status API and the periodic sync sweep.
CREATE INDEX virtual_files_staged_idx ON virtual_files (project_id, environment) WHERE sync_state <> 'synced';

-- Interval for periodic auto-sync of user files to Git. 0 = disabled (default). Configured by project admins.
ALTER TABLE projects ADD COLUMN user_files_sync_interval_seconds BIGINT NOT NULL DEFAULT 0;
-- Time of the last successful user-files sync (manual or scheduled). Used to compute when auto-sync is due.
ALTER TABLE projects ADD COLUMN user_files_synced_on TIMESTAMPTZ;
-- Error from the most recent user-files sync attempt. Empty when the last sync succeeded (or none ran yet).
ALTER TABLE projects ADD COLUMN user_files_sync_error TEXT NOT NULL DEFAULT '';
-- Warning from the most recent successful user-files sync, e.g. direct Git edits that the sync overwrote.
ALTER TABLE projects ADD COLUMN user_files_sync_warning TEXT NOT NULL DEFAULT '';
