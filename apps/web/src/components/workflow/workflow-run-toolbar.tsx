"use client"

import { useEffect, useMemo, useState } from "react"
import { Play, Radio, RefreshCw, Save, Square } from "lucide-react"
import { useRouter, useSearchParams } from "next/navigation"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import {
  buildWorkflowSchema,
  validateWorkflowSchema,
} from "@/lib/workflow-schema"
import { useAuthStore } from "@/stores/auth-store"
import { useLLMStreamStore } from "@/stores/llm-stream-store"
import { useWorkflowDesignerStore } from "@/stores/workflow-designer-store"
import { useWorkflowRunStore } from "@/stores/workflow-run-store"
import { useWorkflowStore } from "@/stores/workflow-store"
import type { LLMStreamRequest } from "@/types/llm-stream"
import type { WorkflowRunStatus } from "@/types/workflow-run"

const DEFAULT_RUN_INPUT = `{
  "message": "你好，请运行当前 Workflow"
}`

export function WorkflowRunToolbar() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const queryWorkflowId =
    searchParams.get("workflowId") ?? searchParams.get("workflow_id") ?? ""

  const currentWorkspace = useAuthStore((state) => state.currentWorkspace)

  const nodes = useWorkflowDesignerStore((state) => state.nodes)
  const edges = useWorkflowDesignerStore((state) => state.edges)

  const workflowStatus = useWorkflowStore((state) => state.status)
  const currentWorkflowWorkspaceId = useWorkflowStore(
    (state) => state.currentWorkspaceId,
  )
  const currentWorkflowId = useWorkflowStore(
    (state) => state.currentWorkflowId,
  )
  const draftName = useWorkflowStore((state) => state.draftName)
  const workflowErrorMessage = useWorkflowStore((state) => state.errorMessage)
  const setDraftName = useWorkflowStore((state) => state.setDraftName)
  const loadWorkflow = useWorkflowStore((state) => state.loadWorkflow)
  const saveWorkflow = useWorkflowStore((state) => state.saveWorkflow)

  const workflowId =
    queryWorkflowId ||
    (currentWorkflowWorkspaceId === currentWorkspace?.id
      ? currentWorkflowId
      : null) ||
    ""

  const requestStatus = useWorkflowRunStore((state) => state.requestStatus)
  const currentRun = useWorkflowRunStore((state) => state.currentRun)
  const startRun = useWorkflowRunStore((state) => state.startRun)
  const refreshRun = useWorkflowRunStore((state) => state.refreshRun)
  const cancelRun = useWorkflowRunStore((state) => state.cancelRun)
  const setErrorMessage = useWorkflowRunStore((state) => state.setErrorMessage)

  const streamStatus = useLLMStreamStore((state) => state.status)
  const startStream = useLLMStreamStore((state) => state.startStream)
  const cancelStream = useLLMStreamStore((state) => state.cancelStream)

  const [runInputText, setRunInputText] = useState(DEFAULT_RUN_INPUT)

  useEffect(() => {
    if (!currentWorkspace?.id || !queryWorkflowId) {
      return
    }

    if (
      currentWorkflowWorkspaceId === currentWorkspace.id &&
      currentWorkflowId === queryWorkflowId
    ) {
      return
    }

    void loadWorkflow({
      workspaceId: currentWorkspace.id,
      workflowId: queryWorkflowId,
    })
  }, [
    currentWorkspace?.id,
    currentWorkflowWorkspaceId,
    currentWorkflowId,
    loadWorkflow,
    queryWorkflowId,
  ])

  const isRunBusy = ["running", "refreshing", "canceling"].includes(requestStatus)
  const isStreamBusy = streamStatus === "connecting" || streamStatus === "streaming"
  const isBusy = isRunBusy || isStreamBusy

  const canRefresh = Boolean(currentWorkspace?.id && currentRun?.id)
  const canCancel = Boolean(
    currentWorkspace?.id &&
      currentRun?.id &&
      currentRun.status !== "succeeded" &&
      currentRun.status !== "failed" &&
      currentRun.status !== "canceled",
  )

  const runStatusLabel = useMemo(
    () => currentRun?.status ?? "idle",
    [currentRun?.status],
  )

  async function handleSave() {
    if (!currentWorkspace?.id) {
      setErrorMessage("请先选择 Workspace")
      return
    }

    const name = draftName.trim()
    if (!name) {
      setErrorMessage("Workflow 名称不能为空")
      return
    }

    const schema = buildWorkflowSchema({ name, nodes, edges })
    const validation = validateWorkflowSchema(schema)
    if (!validation.valid) {
      setErrorMessage(
        `Workflow 校验失败：${validation.errorCount} 个错误，请在右侧“校验”面板中处理。`,
      )
      return
    }

    const saved = await saveWorkflow({
      workspaceId: currentWorkspace.id,
      workflowId,
      name,
      schema,
    })

    if (!saved) {
      return
    }

    router.replace(
      `/dashboard/workflows?workflowId=${encodeURIComponent(saved.id)}`,
    )
  }

  async function handleRun() {
    if (!currentWorkspace?.id) {
      setErrorMessage("请先选择 Workspace")
      return
    }

    if (currentWorkspace.role === "viewer") {
      setErrorMessage("viewer 角色没有保存和运行 Workflow 的权限")
      return
    }

    const runtimeValidationMessage = validateRuntimeNodeConfigs(nodes, edges)
    if (runtimeValidationMessage) {
      setErrorMessage(runtimeValidationMessage)
      return
    }

    const name = draftName.trim()
    if (!name) {
      setErrorMessage("Workflow 名称不能为空")
      return
    }

    const schema = buildWorkflowSchema({ name, nodes, edges })
    const validation = validateWorkflowSchema(schema)
    if (!validation.valid) {
      setErrorMessage(
        `Workflow 校验失败：${validation.errorCount} 个错误，请在右侧“校验”面板中处理。`,
      )
      return
    }

    const parsedInput = parseRunInput(runInputText)
    if (!parsedInput.ok) {
      setErrorMessage(parsedInput.message)
      return
    }

    const saved = await saveWorkflow({
      workspaceId: currentWorkspace.id,
      workflowId,
      name,
      schema,
    })
    if (!saved) {
      return
    }

    if (saved.id !== queryWorkflowId) {
      router.replace(
        `/dashboard/workflows?workflowId=${encodeURIComponent(saved.id)}`,
      )
    }

    await startRun({
      workspaceId: currentWorkspace.id,
      workflowId: saved.id,
      input: parsedInput.value,
      traceId: createTraceId(),
    })
  }

  async function handleStreamRun() {
    if (!currentWorkspace?.id) {
      setErrorMessage("请先选择 Workspace")
      return
    }

    const parsedInput = parseRunInput(runInputText)
    if (!parsedInput.ok) {
      setErrorMessage(parsedInput.message)
      return
    }

    const streamRequest = buildStreamRequestFromDesigner(nodes, parsedInput.value)
    if (!streamRequest.ok) {
      setErrorMessage(streamRequest.message)
      return
    }

    await startStream({
      workspaceId: currentWorkspace.id,
      payload: streamRequest.value,
    })
  }

  async function handleRefresh() {
    if (!currentWorkspace?.id || !currentRun?.id) {
      return
    }

    await refreshRun({
      workspaceId: currentWorkspace.id,
      runId: currentRun.id,
    })
  }

  async function handleCancel() {
    if (!currentWorkspace?.id || !currentRun?.id) {
      return
    }

    await cancelRun({
      workspaceId: currentWorkspace.id,
      runId: currentRun.id,
    })
  }

  return (
    <section className="border-b border-slate-200 bg-white px-4 py-3">
      <div className="flex flex-col gap-3 xl:flex-row xl:items-start xl:justify-between">
        <div className="min-w-0 flex-1 space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <Badge variant={getRunStatusBadgeVariant(currentRun?.status)}>
              {runStatusLabel}
            </Badge>

            <Badge variant="outline">stream: {streamStatus}</Badge>

            {currentRun?.id ? (
              <Badge variant="outline" className="max-w-full truncate">
                run: {currentRun.id}
              </Badge>
            ) : null}

            {workflowId ? (
              <Badge variant="outline" className="max-w-full truncate">
                workflow: {workflowId}
              </Badge>
            ) : (
              <Badge variant="destructive">workflow 未保存</Badge>
            )}
          </div>

          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <Input
              aria-label="Workflow 名称"
              value={draftName}
              maxLength={255}
              className="sm:max-w-sm"
              onChange={(event) => setDraftName(event.target.value)}
            />
            <span className="text-xs text-slate-500">
              {workflowStatus === "saving"
                ? "正在保存..."
                : workflowStatus === "loading"
                  ? "正在加载..."
                  : workflowErrorMessage}
            </span>
          </div>

          <Textarea
            value={runInputText}
            rows={3}
            className="min-h-20 font-mono text-xs"
            onChange={(event) => setRunInputText(event.target.value)}
          />
        </div>

        <div className="flex shrink-0 flex-wrap gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={
              isBusy ||
              workflowStatus === "saving" ||
              workflowStatus === "loading" ||
              currentWorkspace?.role === "viewer"
            }
            onClick={handleSave}
          >
            <Save className="h-4 w-4" />
            {workflowId ? "保存" : "首次保存"}
          </Button>

          <Button
            size="sm"
            disabled={
              isBusy ||
              workflowStatus === "saving" ||
              workflowStatus === "loading" ||
              currentWorkspace?.role === "viewer"
            }
            onClick={handleRun}
          >
            <Play className="h-4 w-4" />
            运行
          </Button>

          <Button
            variant="outline"
            size="sm"
            disabled={isBusy}
            onClick={handleStreamRun}
          >
            <Radio className="h-4 w-4" />
            流式运行
          </Button>

          <Button
            variant="outline"
            size="sm"
            disabled={!isStreamBusy}
            onClick={cancelStream}
          >
            <Square className="h-4 w-4" />
            停止流式
          </Button>

          <Button
            variant="outline"
            size="sm"
            disabled={isBusy || !canRefresh}
            onClick={handleRefresh}
          >
            <RefreshCw className="h-4 w-4" />
            刷新
          </Button>

          <Button
            variant="outline"
            size="sm"
            disabled={isBusy || !canCancel}
            onClick={handleCancel}
          >
            <Square className="h-4 w-4" />
            取消
          </Button>
        </div>
      </div>
    </section>
  )
}

function validateRuntimeNodeConfigs(
  nodes: ReturnType<typeof useWorkflowDesignerStore.getState>["nodes"],
  edges: ReturnType<typeof useWorkflowDesignerStore.getState>["edges"],
) {
  const executableNodeTypes = new Set(["Start", "Prompt", "LLM", "End"])

  for (const node of nodes) {
    if (!executableNodeTypes.has(node.data.nodeType)) {
      return `节点 ${node.data.label || node.id}（${node.data.nodeType}）暂不支持运行。`
    }

    if (
      node.data.nodeType === "Prompt" &&
      !node.data.config.promptTemplate?.trim()
    ) {
      return `Prompt 节点 ${node.data.label || node.id} 缺少 Prompt 模板。请选中该节点，在右侧属性面板填写 promptTemplate。`
    }

    if (node.data.nodeType !== "LLM") {
      continue
    }

    const config = node.data.config
    if (!config.provider?.trim()) {
      return `LLM 节点 ${node.data.label || node.id} 缺少 Provider。请选中该节点，在右侧属性面板填写 provider。`
    }

    if (!config.model?.trim()) {
      return `LLM 节点 ${node.data.label || node.id} 缺少 Model。请选中该节点，在右侧属性面板填写 model。`
    }

    if (
      config.temperature !== undefined &&
      (config.temperature < 0 || config.temperature > 2)
    ) {
      return `LLM 节点 ${node.data.label || node.id} 的 Temperature 必须在 0 到 2 之间。`
    }

    if (config.maxTokens !== undefined && config.maxTokens <= 0) {
      return `LLM 节点 ${node.data.label || node.id} 的 Max Tokens 必须大于 0。`
    }

    const hasPromptInput = edges.some((edge) => {
      if (edge.target !== node.id) {
        return false
      }

      return nodes.some(
        (sourceNode) =>
          sourceNode.id === edge.source &&
          sourceNode.data.nodeType === "Prompt",
      )
    })

    if (!hasPromptInput) {
      return `LLM 节点 ${node.data.label || node.id} 必须连接一个上游 Prompt 节点。`
    }
  }

  return ""
}

function parseRunInput(text: string):
  | {
      ok: true
      value: Record<string, unknown>
    }
  | {
      ok: false
      message: string
    } {
  try {
    const parsed = JSON.parse(text) as unknown

    if (
      parsed === null ||
      Array.isArray(parsed) ||
      typeof parsed !== "object"
    ) {
      return {
        ok: false,
        message: "Run input 必须是 JSON 对象",
      }
    }

    return {
      ok: true,
      value: parsed as Record<string, unknown>,
    }
  } catch (error) {
    return {
      ok: false,
      message: error instanceof Error ? error.message : "Run input JSON 格式错误",
    }
  }
}

function buildStreamRequestFromDesigner(
  nodes: ReturnType<typeof useWorkflowDesignerStore.getState>["nodes"],
  runInput: Record<string, unknown>,
):
  | {
      ok: true
      value: LLMStreamRequest
    }
  | {
      ok: false
      message: string
    } {
  const llmNode = nodes.find((node) => node.data.nodeType === "LLM")
  if (!llmNode) {
    return {
      ok: false,
      message: "当前画布没有 LLM 节点",
    }
  }

  const config = llmNode.data.config
  const provider = normalizeRequiredString(config.provider)
  const model = normalizeRequiredString(config.model)

  if (!provider || !model) {
    return {
      ok: false,
      message: "LLM 节点必须配置 provider 和 model",
    }
  }

  const userPrompt = buildStreamUserPrompt(nodes, runInput)
  if (!userPrompt) {
    return {
      ok: false,
      message: "没有可用于流式运行的 Prompt 或 message",
    }
  }

  const messages: LLMStreamRequest["messages"] = []

  if (config.systemPrompt?.trim()) {
    messages.push({
      role: "system",
      content: config.systemPrompt.trim(),
    })
  }

  messages.push({
    role: "user",
    content: userPrompt,
  })

  return {
    ok: true,
    value: {
      provider,
      model,
      messages,
      temperature: config.temperature,
      max_tokens: config.maxTokens,
      metadata: {
        trace_id: createTraceId(),
        node_id: llmNode.id,
        node_type: "LLM",
        source: "workflow_designer_stream_panel",
      },
    },
  }
}

function buildStreamUserPrompt(
  nodes: ReturnType<typeof useWorkflowDesignerStore.getState>["nodes"],
  runInput: Record<string, unknown>,
) {
  const promptNode = nodes.find((node) => node.data.nodeType === "Prompt")
  const promptTemplate = promptNode?.data.config.promptTemplate?.trim()

  if (promptTemplate) {
    return renderSimplePromptTemplate(promptTemplate, runInput)
  }

  const message = runInput.message
  if (typeof message === "string" && message.trim()) {
    return message.trim()
  }

  return JSON.stringify(runInput, null, 2)
}

function renderSimplePromptTemplate(
  template: string,
  input: Record<string, unknown>,
) {
  return template.replace(/\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}/g, (_, key: string) => {
    const value = readPath(input, key)

    if (value === undefined || value === null) {
      return ""
    }

    if (typeof value === "string") {
      return value
    }

    return JSON.stringify(value)
  })
}

function readPath(value: Record<string, unknown>, path: string) {
  return path.split(".").reduce<unknown>((current, key) => {
    if (
      current === null ||
      Array.isArray(current) ||
      typeof current !== "object"
    ) {
      return undefined
    }

    return (current as Record<string, unknown>)[key]
  }, value)
}

function normalizeRequiredString(value: unknown) {
  if (typeof value !== "string") {
    return ""
  }

  return value.trim()
}

function createTraceId() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return `web_${crypto.randomUUID()}`
  }

  return `web_${Date.now()}`
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
