"use client"

import { AlertCircle, Bot, CheckCircle2, Clock3, Timer } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useWorkflowRunStore } from "@/stores/workflow-run-store"
import type {
  NodeExecutionStatus,
  WorkflowRunNodeExecution,
  WorkflowRunStatus,
  WorkflowRunTokenUsage,
} from "@/types/workflow-run"

export function WorkflowRunPanel() {
  const requestStatus = useWorkflowRunStore((state) => state.requestStatus)
  const currentRun = useWorkflowRunStore((state) => state.currentRun)
  const nodeExecutions = useWorkflowRunStore((state) => state.nodeExecutions)
  const errorMessage = useWorkflowRunStore((state) => state.errorMessage)

  const llmNodes = nodeExecutions.filter((node) => node.node_type === "LLM")

  return (
    <section className="min-h-72 border-t border-slate-200 bg-white">
      <div className="flex h-12 items-center justify-between border-b border-slate-200 px-4">
        <div className="flex items-center gap-2">
          <Timer className="h-4 w-4 text-slate-500" />
          <h2 className="text-sm font-semibold text-slate-950">运行结果</h2>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="outline">{requestStatus}</Badge>
          {currentRun ? (
            <Badge variant={getRunStatusBadgeVariant(currentRun.status)}>
              {currentRun.status}
            </Badge>
          ) : null}
        </div>
      </div>

      <ScrollArea className="h-72">
        <div className="space-y-4 p-4">
          {errorMessage ? (
            <div className="rounded-md border border-rose-200 bg-rose-50 p-3">
              <div className="flex items-start gap-2">
                <AlertCircle className="mt-0.5 h-4 w-4 shrink-0 text-rose-600" />
                <p className="text-sm text-rose-800">{errorMessage}</p>
              </div>
            </div>
          ) : null}

          {!currentRun && !errorMessage ? (
            <EmptyRunState />
          ) : null}

          {currentRun ? (
            <RunSummary run={currentRun} nodeCount={nodeExecutions.length} />
          ) : null}

          {llmNodes.length > 0 ? (
            <section className="space-y-2">
              <h3 className="text-xs font-semibold uppercase text-slate-500">
                LLM 输出
              </h3>

              <div className="space-y-2">
                {llmNodes.map((node) => (
                  <LLMOutputCard key={node.id} node={node} />
                ))}
              </div>
            </section>
          ) : null}

          {nodeExecutions.length > 0 ? (
            <section className="space-y-2">
              <h3 className="text-xs font-semibold uppercase text-slate-500">
                节点执行
              </h3>

              <div className="space-y-2">
                {nodeExecutions.map((node) => (
                  <NodeExecutionRow key={node.id} node={node} />
                ))}
              </div>
            </section>
          ) : null}
        </div>
      </ScrollArea>
    </section>
  )
}

function EmptyRunState() {
  return (
    <div className="flex min-h-36 items-center justify-center rounded-md border border-dashed border-slate-200 bg-slate-50">
      <div className="text-center">
        <Clock3 className="mx-auto h-5 w-5 text-slate-400" />
        <p className="mt-2 text-sm font-medium text-slate-700">暂无运行记录</p>
      </div>
    </div>
  )
}

function RunSummary({
  run,
  nodeCount,
}: {
  run: {
    id: string
    status: WorkflowRunStatus
    output?: unknown
    started_at?: string | null
    finished_at?: string | null
  }
  nodeCount: number
}) {
  return (
    <div className="rounded-md border border-slate-200 bg-slate-50 p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-slate-950">
            {run.id}
          </p>
          <p className="mt-1 text-xs text-slate-500">
            {formatRunTime(run.started_at)} - {formatRunTime(run.finished_at)}
          </p>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="outline">{nodeCount} nodes</Badge>
          <Badge variant={getRunStatusBadgeVariant(run.status)}>
            {run.status}
          </Badge>
        </div>
      </div>

      {run.output ? (
        <pre className="mt-3 max-h-28 overflow-auto rounded-md bg-slate-950 p-3 text-xs leading-5 text-slate-50">
          {formatJSON(run.output)}
        </pre>
      ) : null}
    </div>
  )
}

function LLMOutputCard({ node }: { node: WorkflowRunNodeExecution }) {
  const responseText = readResponseText(node.output)
  const tokenUsage = readTokenUsage(node.token_usage)

  return (
    <div className="rounded-md border border-slate-200 bg-white p-3">
      <div className="flex items-start gap-3">
        <div className="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-slate-100 text-slate-600">
          <Bot className="h-4 w-4" />
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="text-sm font-semibold text-slate-950">
              {node.node_id}
            </p>
            <Badge variant={getNodeStatusBadgeVariant(node.status)}>
              {node.status}
            </Badge>
            <Badge variant="outline">{node.latency_ms} ms</Badge>
          </div>

          {responseText ? (
            <p className="mt-2 whitespace-pre-wrap text-sm leading-6 text-slate-700">
              {responseText}
            </p>
          ) : (
            <p className="mt-2 text-sm text-slate-500">暂无 response_text</p>
          )}

          {tokenUsage ? (
            <div className="mt-3 flex flex-wrap gap-2">
              <Badge variant="outline">
                input {tokenUsage.input_tokens ?? 0}
              </Badge>
              <Badge variant="outline">
                output {tokenUsage.output_tokens ?? 0}
              </Badge>
              <Badge variant="outline">
                total {tokenUsage.total_tokens ?? 0}
              </Badge>
            </div>
          ) : null}
        </div>
      </div>
    </div>
  )
}

function NodeExecutionRow({ node }: { node: WorkflowRunNodeExecution }) {
  const tokenUsage = readTokenUsage(node.token_usage)

  return (
    <div className="rounded-md border border-slate-200 bg-white p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            {node.status === "succeeded" ? (
              <CheckCircle2 className="h-4 w-4 text-emerald-600" />
            ) : node.status === "failed" ? (
              <AlertCircle className="h-4 w-4 text-rose-600" />
            ) : (
              <Clock3 className="h-4 w-4 text-slate-500" />
            )}

            <p className="text-sm font-medium text-slate-950">
              {node.node_id}
            </p>

            <Badge variant="outline">{node.node_type}</Badge>
          </div>

          <p className="mt-1 text-xs text-slate-500">
            sequence {node.sequence}
          </p>
        </div>

        <div className="flex shrink-0 flex-wrap gap-2">
          <Badge variant={getNodeStatusBadgeVariant(node.status)}>
            {node.status}
          </Badge>
          <Badge variant="outline">{node.latency_ms} ms</Badge>
          {tokenUsage ? (
            <Badge variant="outline">
              token {tokenUsage.total_tokens ?? 0}
            </Badge>
          ) : null}
        </div>
      </div>

      {node.error ? (
        <pre className="mt-3 max-h-24 overflow-auto rounded-md bg-rose-950 p-3 text-xs leading-5 text-rose-50">
          {formatJSON(node.error)}
        </pre>
      ) : null}
    </div>
  )
}

function getRunStatusBadgeVariant(status?: WorkflowRunStatus) {
  switch (status) {
    case "succeeded":
      return "secondary"

    case "failed":
    case "canceled":
      return "destructive"

    case "pending":
    case "running":
      return "outline"

    default:
      return "outline"
  }
}

function getNodeStatusBadgeVariant(status: NodeExecutionStatus) {
  switch (status) {
    case "succeeded":
      return "secondary"

    case "failed":
      return "destructive"

    default:
      return "outline"
  }
}

function readResponseText(output: unknown) {
  const value = asRecord(output)
  if (!value) {
    return ""
  }

  if (typeof value.response_text === "string") {
    return value.response_text
  }

  if (typeof value.text === "string") {
    return value.text
  }

  return ""
}

function readTokenUsage(value: unknown): WorkflowRunTokenUsage | null {
  const record = asRecord(value)
  if (!record) {
    return null
  }

  return record as WorkflowRunTokenUsage
}

function asRecord(value: unknown): Record<string, unknown> | null {
  if (value === null || Array.isArray(value) || typeof value !== "object") {
    return null
  }

  return value as Record<string, unknown>
}

function formatJSON(value: unknown) {
  return JSON.stringify(value, null, 2)
}

function formatRunTime(value?: string | null) {
  if (!value) {
    return "--"
  }

  return new Date(value).toLocaleString()
}