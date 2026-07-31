CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    schema_version VARCHAR(32) NOT NULL DEFAULT '1.0',
    schema_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_workflows_workspace
        FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_workflows_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_workflows_updated_by
        FOREIGN KEY (updated_by)
        REFERENCES users(id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_workflows_workspace_id
ON workflows(workspace_id);

CREATE INDEX IF NOT EXISTS idx_workflows_deleted_at
ON workflows(deleted_at);

CREATE INDEX IF NOT EXISTS idx_workflows_workspace_updated_at
ON workflows(workspace_id, updated_at DESC);