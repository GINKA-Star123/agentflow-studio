import type { WorkflowSchema } from "@/lib/workflow-schema"

export type WorkflowDefinitionSummary = {
  id: string
  workspace_id: string
  name: string
  schema_version: string
  node_count: number
  edge_count: number
  created_at: string
  updated_at: string
}

export type WorkflowDefinition = WorkflowDefinitionSummary & {
  schema: WorkflowSchema
}

export type WorkflowListResult = {
  items: WorkflowDefinitionSummary[]
}

export type SaveWorkflowPayload = {
  name: string
  schema: WorkflowSchema
}
