-- Users of embedded dashboards, identified by the external user ID passed to GetIFrame.
-- The ID is the hash of the external user ID and project ID, which is also the subject of the embed's runtime JWT.
-- A row stores the identity the embed user had in their latest iframe, which reports created from an embed run as.
CREATE TABLE embed_users (
    id TEXT PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    external_user_id TEXT NOT NULL,
    email TEXT,
    attributes JSONB,
    last_seen_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_on TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_on TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX embed_users_project_id_external_user_id_idx ON embed_users (project_id, external_user_id);

-- Virtual files owned by an embed user, such as reports created from an embed.
-- Deleting an embed user must soft-delete their virtual files first, so the runtime learns about the deletion.
ALTER TABLE virtual_files ADD COLUMN embed_owner_id TEXT REFERENCES embed_users (id) ON DELETE SET NULL;
CREATE INDEX virtual_files_embed_owner_id_idx ON virtual_files (embed_owner_id);
