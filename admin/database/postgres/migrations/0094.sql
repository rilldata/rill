ALTER TABLE project_roles ADD create_personal_canvases BOOLEAN NOT NULL DEFAULT false;
UPDATE project_roles SET create_personal_canvases = read_prod;
