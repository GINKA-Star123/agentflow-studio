CREATE TABLE IF NOT EXISTS workflow_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    workspace_id UUID NOT NULL,
    workflow_id UUID NOT NULL,
    triggered_by UUID NOT NULL,

    status VARCHAR(32) NOT NULL DEFAULT 'pending',

    schema_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    input JSONB NOT NULL DEFAULT '{}'::jsonb,
    output JSONB,
    error JSONB,

    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_workflow_runs_workspace
        FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_workflow_runs_workflow
        FOREIGN KEY (workflow_id)
        REFERENCES workflows(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_workflow_runs_triggered_by
        FOREIGN KEY (triggered_by)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT workflow_runs_status_check
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'canceled'))
);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_workspace_id
ON workflow_runs(workspace_id);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow_id
ON workflow_runs(workflow_id);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_triggered_by
ON workflow_runs(triggered_by);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_status
ON workflow_runs(status);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_workspace_workflow_created_at
ON workflow_runs(workspace_id, workflow_id, created_at DESC);


CREATE TABLE IF NOT EXISTS node_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    workspace_id UUID NOT NULL,
    run_id UUID NOT NULL,

    node_id VARCHAR(128) NOT NULL,
    node_type VARCHAR(64) NOT NULL,
    sequence INTEGER NOT NULL DEFAULT 0,

    status VARCHAR(32) NOT NULL DEFAULT 'pending',

    input JSONB NOT NULL DEFAULT '{}'::jsonb,
    output JSONB,
    error JSONB,
    token_usage JSONB,

    latency_ms BIGINT NOT NULL DEFAULT 0,

    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_node_executions_workspace
        FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_node_executions_run
        FOREIGN KEY (run_id)
        REFERENCES workflow_runs(id)
        ON DELETE CASCADE,

    CONSTRAINT node_executions_status_check
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'skipped')),

    CONSTRAINT node_executions_sequence_check
        CHECK (sequence >= 0)
);

CREATE INDEX IF NOT EXISTS idx_node_executions_workspace_id
ON node_executions(workspace_id);

CREATE INDEX IF NOT EXISTS idx_node_executions_run_id
ON node_executions(run_id);

CREATE INDEX IF NOT EXISTS idx_node_executions_status
ON node_executions(status);

CREATE INDEX IF NOT EXISTS idx_node_executions_node
ON node_executions(run_id, node_id);

CREATE INDEX IF NOT EXISTS idx_node_executions_workspace_run
ON node_executions(workspace_id, run_id, created_at ASC);