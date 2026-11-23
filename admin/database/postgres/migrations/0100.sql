-- Bookmarks owned by an embed user, created from an embed.
-- Exactly one of user_id and embed_owner_id is set.
-- Bookmarks have no runtime state, so they can be deleted together with their embed user.
ALTER TABLE bookmarks ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE bookmarks ADD COLUMN embed_owner_id TEXT REFERENCES embed_users (id) ON DELETE CASCADE;
ALTER TABLE bookmarks ADD CONSTRAINT bookmarks_owner_check CHECK ((user_id IS NULL) <> (embed_owner_id IS NULL));
CREATE INDEX bookmarks_embed_owner_id_idx ON bookmarks (embed_owner_id);
