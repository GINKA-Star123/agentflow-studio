import type {
  WorkflowDesignerEdge,
  WorkflowDesignerNode,
  WorkflowNodeConfig,
  WorkflowNodeType,
} from "@/types/workflow"

export const WORKFLOW_SCHEMA_VERSION = "1.0" as const
export const DEFAULT_WORKFLOW_NAME = "未命名 Workflow" as const

export type WorkflowSchema = {
  schema_version: typeof WORKFLOW_SCHEMA_VERSION
  name: string
  summary: {
    node_count: number
    edge_count: number
    start_count: number
    end_count: number
  }
  nodes: WorkflowSchemaNode[]
  edges: WorkflowSchemaEdge[]
}

export type WorkflowSchemaNode = {
  id: string
  type: WorkflowNodeType
  label: string
  description: string
  position: {
    x: number
    y: number
  }
  config: WorkflowNodeConfig
}

export type WorkflowSchemaEdge = {
  id: string
  source: string
  target: string
  sourceHandle?: string | null
  targetHandle?: string | null
  type?: string | null
}

export type WorkflowValidationSeverity = "error" | "warning"

export type WorkflowValidationCode =
  | "START_COUNT_INVALID"
  | "END_COUNT_INVALID"
  | "EDGE_SOURCE_MISSING"
  | "EDGE_TARGET_MISSING"
  | "EDGE_FROM_END"
  | "EDGE_TO_START"
  | "SELF_LOOP"
  | "ISOLATED_NODE"
  | "DUPLICATE_EDGE"

export type WorkflowValidationIssue = {
  id: string
  severity: WorkflowValidationSeverity
  code: WorkflowValidationCode
  message: string
  nodeId?: string
  edgeId?: string
}

export type WorkflowValidationResult = {
  valid: boolean
  errorCount: number
  warningCount: number
  issues: WorkflowValidationIssue[]
}

export function buildWorkflowSchema(input: {
  name: string
  nodes: WorkflowDesignerNode[]
  edges: WorkflowDesignerEdge[]
}): WorkflowSchema {
  const nodes = input.nodes.map((node) => ({
    id: node.id,
    type: node.data.nodeType,
    label: node.data.label,
    description: node.data.description,
    position: {
      x: roundCoordinate(node.position.x),
      y: roundCoordinate(node.position.y),
    },
    config: {
      ...(node.data.config ?? {}),
    },
  }))

  const edges = input.edges.map((edge) => ({
    id: edge.id,
    source: edge.source,
    target: edge.target,
    sourceHandle: edge.sourceHandle ?? null,
    targetHandle: edge.targetHandle ?? null,
    type: edge.type ?? null,
  }))

  const startCount = nodes.filter((node) => node.type === "Start").length
  const endCount = nodes.filter((node) => node.type === "End").length

  return {
    schema_version: WORKFLOW_SCHEMA_VERSION,
    name: input.name.trim() || DEFAULT_WORKFLOW_NAME,
    summary: {
      node_count: nodes.length,
      edge_count: edges.length,
      start_count: startCount,
      end_count: endCount,
    },
    nodes,
    edges,
  }
}

export function validateWorkflowSchema(
  schema: WorkflowSchema,
): WorkflowValidationResult {
  const issues: WorkflowValidationIssue[] = []
  const nodeMap = new Map(schema.nodes.map((node) => [node.id, node]))
  const incidentNodeIds = new Set<string>()
  const edgeSignatureMap = new Map<string, { edgeId: string; count: number }>()

  if (schema.summary.start_count !== 1) {
    issues.push(
      createIssue(
        "error",
        "START_COUNT_INVALID",
        `Start 节点必须且只能有 1 个，当前为 ${schema.summary.start_count} 个。`,
      ),
    )
  }

  if (schema.summary.end_count !== 1) {
    issues.push(
      createIssue(
        "error",
        "END_COUNT_INVALID",
        `End 节点必须且只能有 1 个，当前为 ${schema.summary.end_count} 个。`,
      ),
    )
  }

  for (const edge of schema.edges) {
    const sourceNode = nodeMap.get(edge.source) ?? null
    const targetNode = nodeMap.get(edge.target) ?? null

    const signature = createEdgeSignature(edge)
    const signatureEntry = edgeSignatureMap.get(signature)

    if (signatureEntry) {
      signatureEntry.count += 1
    } else {
      edgeSignatureMap.set(signature, {
        edgeId: edge.id,
        count: 1,
      })
    }

    if (sourceNode) {
      incidentNodeIds.add(sourceNode.id)
    } else {
      issues.push(
        createIssue(
          "error",
          "EDGE_SOURCE_MISSING",
          `连线 ${edge.id} 的来源节点 ${edge.source} 不存在。`,
          { edgeId: edge.id },
        ),
      )
    }

    if (targetNode) {
      incidentNodeIds.add(targetNode.id)
    } else {
      issues.push(
        createIssue(
          "error",
          "EDGE_TARGET_MISSING",
          `连线 ${edge.id} 的目标节点 ${edge.target} 不存在。`,
          { edgeId: edge.id },
        ),
      )
    }

    if (!sourceNode || !targetNode) {
      continue
    }

    if (edge.source === edge.target) {
      issues.push(
        createIssue(
          "error",
          "SELF_LOOP",
          `连线 ${edge.id} 不能连接到自身。`,
          { edgeId: edge.id },
        ),
      )
      continue
    }

    if (sourceNode.type === "End") {
      issues.push(
        createIssue(
          "error",
          "EDGE_FROM_END",
          `End 节点 ${sourceNode.id} 不能作为连线来源。`,
          { nodeId: sourceNode.id, edgeId: edge.id },
        ),
      )
    }

    if (targetNode.type === "Start") {
      issues.push(
        createIssue(
          "error",
          "EDGE_TO_START",
          `Start 节点 ${targetNode.id} 不能作为连线目标。`,
          { nodeId: targetNode.id, edgeId: edge.id },
        ),
      )
    }
  }

  for (const node of schema.nodes) {
    if (!incidentNodeIds.has(node.id)) {
      issues.push(
        createIssue(
          "warning",
          "ISOLATED_NODE",
          `节点 ${node.label || node.id} 是孤立节点，没有连接任何连线。`,
          { nodeId: node.id },
        ),
      )
    }
  }

  for (const [signature, entry] of edgeSignatureMap.entries()) {
    if (entry.count > 1) {
      issues.push(
        createIssue(
          "warning",
          "DUPLICATE_EDGE",
          `存在重复连线：${signature}。`,
          { edgeId: entry.edgeId },
        ),
      )
    }
  }

  const errorCount = issues.filter((issue) => issue.severity === "error").length
  const warningCount = issues.filter(
    (issue) => issue.severity === "warning",
  ).length

  return {
    valid: errorCount === 0,
    errorCount,
    warningCount,
    issues,
  }
}

export function formatWorkflowSchema(schema: WorkflowSchema) {
  return JSON.stringify(schema, null, 2)
}

function roundCoordinate(value: number) {
  return Math.round(value * 100) / 100
}

function createEdgeSignature(edge: WorkflowSchemaEdge) {
  return [
    edge.source,
    edge.sourceHandle ?? "default",
    edge.target,
    edge.targetHandle ?? "default",
  ].join("::")
}

function createIssue(
  severity: WorkflowValidationSeverity,
  code: WorkflowValidationCode,
  message: string,
  extra?: Pick<WorkflowValidationIssue, "nodeId" | "edgeId">,
): WorkflowValidationIssue {
  const scope = extra?.nodeId ?? extra?.edgeId ?? "global"

  return {
    id: `${severity}-${code}-${scope}`,
    severity,
    code,
    message,
    ...extra,
  }
}