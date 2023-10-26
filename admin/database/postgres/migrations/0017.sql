CREATE TABLE virtual_files (
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    branch TEXT NOT NULL,
    path TEXT NOT NULL,
    data BYTEA NOT NULL,
    deleted BOOLEAN NOT NULL,
    created_on TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_on TIMESTAMPTZ DEFAULT now() NOT NULL,
    PRIMARY KEY (project_id, branch, path)
);

CREATE INDEX virtual_files_project_id_branch_updated_on_idx ON virtual_files (project_id, branch, updated_on);
