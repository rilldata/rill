-- Update the 'editor' project role to allow it to create and manage magic auth tokens.
UPDATE project_roles SET create_magic_auth_tokens = true WHERE name = 'editor';
UPDATE project_roles SET manage_magic_auth_tokens = true WHERE name = 'editor';
