export type WorkflowRunStatus =
  | "pending"
  | "running"
  | "succeeded"
  | "failed"
  | "canceled"

export type NodeExecutionStatus =
  | "pending"
  | "running"
  | "succeeded"
  | "failed"
  | "skipped"

export type WorkflowRunTokenUsage = {
  provider?: string
  model?: string
  input_tokens?: number
  output_tokens?: number
  total_tokens?: number
}

export type WorkflowRunDetail = {
  id: string
  workspace_id: string
  workflow_id: string
  triggered_by?: string
  status: WorkflowRunStatus

  input?: unknown
  output?: unknown
  error?: unknown

  started_at?: string | null
  finished_at?: string | null
  created_at?: string
  updated_at?: string
}

export type WorkflowRunNodeExecution = {
  id: string
  workspace_id: string
  run_id: string

  node_id: string
  node_type: string
  sequence: number
  status: NodeExecutionStatus

  input?: unknown
  output?: unknown
  error?: unknown
  token_usage?: WorkflowRunTokenUsage | null
  latency_ms: number

  started_at?: string | null
  finished_at?: string | null
  created_at?: string
  updated_at?: string
}

export type WorkflowRunNodeExecutionListResult = {
  items: WorkflowRunNodeExecution[]
}

export type StartWorkflowRunPayload = {
  input?: Record<string, unknown>
  trace_id?: string
}

export type CancelWorkflowRunResult = {
  canceled: boolean
  run: WorkflowRunDetail
}