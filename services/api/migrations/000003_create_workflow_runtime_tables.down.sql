DROP INDEX IF EXISTS idx_node_executions_workspace_run;
DROP INDEX IF EXISTS idx_node_executions_node;
DROP INDEX IF EXISTS idx_node_executions_status;
DROP INDEX IF EXISTS idx_node_executions_run_id;
DROP INDEX IF EXISTS idx_node_executions_workspace_id;
DROP TABLE IF EXISTS node_executions;

DROP INDEX IF EXISTS idx_workflow_runs_workspace_workflow_created_at;
DROP INDEX IF EXISTS idx_workflow_runs_status;
DROP INDEX IF EXISTS idx_workflow_runs_triggered_by;
DROP INDEX IF EXISTS idx_workflow_runs_workflow_id;
DROP INDEX IF EXISTS idx_workflow_runs_workspace_id;
DROP TABLE IF EXISTS workflow_runs;