"use client"

import {
  useMemo,
  useState,
} from "react"
import {
  AlertCircle,
  Cable,
  CheckCircle2,
  Code2,
  Copy,
  MousePointer2,
  Settings2,
  SlidersHorizontal,
  TriangleAlert,
  Trash2,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import {
  DEFAULT_WORKFLOW_NAME,
  buildWorkflowSchema,
  formatWorkflowSchema,
  validateWorkflowSchema,
  type WorkflowSchema,
  type WorkflowValidationIssue,
  type WorkflowValidationResult,
} from "@/lib/workflow-schema"
import { useWorkflowDesignerStore } from "@/stores/workflow-designer-store"
import { useWorkflowStore } from "@/stores/workflow-store"

export function PropertiesPanel() {
  const nodes = useWorkflowDesignerStore((state) => state.nodes)
  const edges = useWorkflowDesignerStore((state) => state.edges)
  const selectedNodeId = useWorkflowDesignerStore((state) => state.selectedNodeId)
  const selectedEdgeId = useWorkflowDesignerStore((state) => state.selectedEdgeId)
  const draftName = useWorkflowStore((state) => state.draftName)
  const deleteNode = useWorkflowDesignerStore((state) => state.deleteNode)
  const deleteEdge = useWorkflowDesignerStore((state) => state.deleteEdge)
  const updateNodeData = useWorkflowDesignerStore((state) => state.updateNodeData)
  const updateNodeConfig = useWorkflowDesignerStore(
    (state) => state.updateNodeConfig,
  )

  const schema = useMemo(
    () =>
      buildWorkflowSchema({
        name: draftName || DEFAULT_WORKFLOW_NAME,
        nodes,
        edges,
      }),
    [draftName, nodes, edges],
  )

  const validation = useMemo(() => validateWorkflowSchema(schema), [schema])
  const schemaJson = useMemo(() => formatWorkflowSchema(schema), [schema])
  const [copiedSchema, setCopiedSchema] = useState("")
  const copied = copiedSchema === schemaJson

  const selectedNode = nodes.find((node) => node.id === selectedNodeId) ?? null
  const selectedEdge = edges.find((edge) => edge.id === selectedEdgeId) ?? null

  async function handleCopySchema() {
    try {
      await navigator.clipboard.writeText(schemaJson)
      setCopiedSchema(schemaJson)
    } catch {
      setCopiedSchema("")
    }
  }

  return (
    <aside className="flex min-h-0 w-full flex-col border-t border-slate-200 bg-white xl:w-80 xl:border-l xl:border-t-0">
      <div className="border-b border-slate-200 px-4 py-4">
        <h2 className="text-sm font-semibold text-slate-950">属性面板</h2>
        <p className="mt-1 text-xs text-slate-500">节点编辑、基础校验与 JSON 预览</p>
      </div>

      <Tabs defaultValue="properties" className="flex min-h-0 flex-1 flex-col">
        <div className="border-b border-slate-200 px-4 py-3">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="properties">属性</TabsTrigger>
            <TabsTrigger value="validation">校验</TabsTrigger>
            <TabsTrigger value="schema">Schema</TabsTrigger>
          </TabsList>
        </div>

        <ScrollArea className="flex-1">
          <TabsContent value="properties" className="space-y-4 p-4">
            {selectedNode ? (
              <NodePropertiesSection
                node={selectedNode}
                deleteNode={deleteNode}
                updateNodeData={updateNodeData}
                updateNodeConfig={updateNodeConfig}
              />
            ) : selectedEdge ? (
              <EdgePropertiesSection edge={selectedEdge} deleteEdge={deleteEdge} />
            ) : (
              <EmptySelectionState />
            )}
          </TabsContent>

          <TabsContent value="validation" className="space-y-4 p-4">
            <ValidationSummaryCard schema={schema} validation={validation} />
            <ValidationIssueList issues={validation.issues} />
          </TabsContent>

          <TabsContent value="schema" className="space-y-4 p-4">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0 space-y-2">
                <div className="flex items-center gap-2">
                  <Code2 className="h-4 w-4 text-slate-500" />
                  <h3 className="text-sm font-semibold text-slate-950">Workflow JSON</h3>
                </div>

                <div className="flex flex-wrap gap-2">
                  <Badge variant="secondary">{schema.schema_version}</Badge>
                  <Badge variant="outline">{schema.summary.node_count} nodes</Badge>
                  <Badge variant="outline">{schema.summary.edge_count} edges</Badge>
                  <Badge variant="outline">{schema.summary.start_count} Start</Badge>
                  <Badge variant="outline">{schema.summary.end_count} End</Badge>
                </div>
              </div>

              <Button variant="outline" size="sm" onClick={handleCopySchema}>
                <Copy className="h-4 w-4" />
                {copied ? "已复制" : "复制 JSON"}
              </Button>
            </div>

            <div className="rounded-md border border-slate-200 bg-slate-950 p-3">
              <pre className="max-h-[460px] overflow-auto whitespace-pre-wrap break-all text-xs leading-5 text-slate-50">
                {schemaJson}
              </pre>
            </div>
          </TabsContent>
        </ScrollArea>
      </Tabs>
    </aside>
  )
}

function NodePropertiesSection({
  node,
  deleteNode,
  updateNodeData,
  updateNodeConfig,
}: {
  node: NonNullable<ReturnType<typeof useWorkflowDesignerStore.getState>["nodes"][number]>
  deleteNode: (nodeId: string) => void
  updateNodeData: (
    nodeId: string,
    patch: {
      label?: string
      description?: string
    },
  ) => void
  updateNodeConfig: (
    nodeId: string,
    patch: {
      promptTemplate?: string
      variables?: string
      provider?: string
      model?: string
      temperature?: number
      maxTokens?: number
      systemPrompt?: string
    },
  ) => void
}) {
  return (
    <div className="space-y-4">
      <div className="flex items-start gap-3">
        <div className="grid h-10 w-10 shrink-0 place-items-center rounded-md bg-slate-100 text-slate-600">
          <MousePointer2 className="h-5 w-5" />
        </div>

        <div className="min-w-0 flex-1">
          <div className="truncate text-sm font-semibold text-slate-950">
            {node.data.label || "未命名节点"}
          </div>
          <div className="mt-1 text-xs text-slate-500">
            {node.data.description || "暂无描述"}
          </div>
        </div>

        <Badge variant="secondary">{node.data.nodeType}</Badge>
      </div>

      <Separator />

      <div className="space-y-3">
        <EditableTextField
          id={`node-label-${node.id}`}
          label="节点名称"
          value={node.data.label}
          placeholder="请输入节点名称"
          onChange={(value) =>
            updateNodeData(node.id, {
              label: value,
            })
          }
        />

        <EditableTextareaField
          id={`node-description-${node.id}`}
          label="节点描述"
          value={node.data.description}
          placeholder="请输入节点描述"
          rows={3}
          onChange={(value) =>
            updateNodeData(node.id, {
              description: value,
            })
          }
        />
      </div>

      <Separator />

      <dl className="space-y-3">
        <ReadonlyPropertyRow label="节点 ID" value={node.id} />
        <ReadonlyPropertyRow label="节点类型" value={node.data.nodeType} />
        <ReadonlyPropertyRow
          label="位置 X"
          value={String(Math.round(node.position.x))}
        />
        <ReadonlyPropertyRow
          label="位置 Y"
          value={String(Math.round(node.position.y))}
        />
      </dl>

      {node.data.nodeType === "Prompt" ? (
        <>
          <Separator />
          <PromptConfigSection
            nodeId={node.id}
            promptTemplate={node.data.config.promptTemplate ?? ""}
            variables={node.data.config.variables ?? ""}
            onChange={updateNodeConfig}
          />
        </>
      ) : null}

      {node.data.nodeType === "LLM" ? (
        <>
          <Separator />
          <LLMConfigSection
            nodeId={node.id}
            provider={node.data.config.provider ?? ""}
            model={node.data.config.model ?? ""}
            temperature={node.data.config.temperature}
            maxTokens={node.data.config.maxTokens}
            systemPrompt={node.data.config.systemPrompt ?? ""}
            onChange={updateNodeConfig}
          />
        </>
      ) : null}

      <Separator />

      <Button
        variant="destructive"
        size="sm"
        className="w-full"
        onClick={() => deleteNode(node.id)}
      >
        <Trash2 className="h-4 w-4" />
        删除节点
      </Button>
    </div>
  )
}

function EdgePropertiesSection({
  edge,
  deleteEdge,
}: {
  edge: NonNullable<ReturnType<typeof useWorkflowDesignerStore.getState>["edges"][number]>
  deleteEdge: (edgeId: string) => void
}) {
  return (
    <div className="space-y-4">
      <div className="flex items-start gap-3">
        <div className="grid h-10 w-10 shrink-0 place-items-center rounded-md bg-slate-100 text-slate-600">
          <Cable className="h-5 w-5" />
        </div>

        <div className="min-w-0 flex-1">
          <div className="truncate text-sm font-semibold text-slate-950">连线</div>
          <div className="mt-1 text-xs text-slate-500">
            {edge.source} -&gt; {edge.target}
          </div>
        </div>

        <Badge variant="secondary">Edge</Badge>
      </div>

      <Separator />

      <dl className="space-y-3">
        <ReadonlyPropertyRow label="连线 ID" value={edge.id} />
        <ReadonlyPropertyRow label="来源节点" value={edge.source} />
        {edge.sourceHandle ? (
          <ReadonlyPropertyRow label="来源句柄" value={edge.sourceHandle} />
        ) : null}
        <ReadonlyPropertyRow label="目标节点" value={edge.target} />
        {edge.targetHandle ? (
          <ReadonlyPropertyRow label="目标句柄" value={edge.targetHandle} />
        ) : null}
        <ReadonlyPropertyRow label="连线类型" value={edge.type ?? "default"} />
      </dl>

      <Separator />

      <Button
        variant="destructive"
        size="sm"
        className="w-full"
        onClick={() => deleteEdge(edge.id)}
      >
        <Trash2 className="h-4 w-4" />
        删除连线
      </Button>
    </div>
  )
}

function PromptConfigSection({
  nodeId,
  promptTemplate,
  variables,
  onChange,
}: {
  nodeId: string
  promptTemplate: string
  variables: string
  onChange: (
    nodeId: string,
    patch: {
      promptTemplate?: string
      variables?: string
    },
  ) => void
}) {
  return (
    <section className="space-y-3">
      <PanelSectionTitle title="Prompt 配置" />

      <EditableTextareaField
        id={`prompt-template-${nodeId}`}
        label="Prompt 模板"
        value={promptTemplate}
        placeholder="例如：请根据 {{input}} 生成一段总结。"
        rows={5}
        onChange={(value) =>
          onChange(nodeId, {
            promptTemplate: value,
          })
        }
      />

      <EditableTextField
        id={`prompt-variables-${nodeId}`}
        label="变量预留"
        value={variables}
        placeholder="例如：input, topic, language"
        onChange={(value) =>
          onChange(nodeId, {
            variables: value,
          })
        }
      />
    </section>
  )
}

function LLMConfigSection({
  nodeId,
  provider,
  model,
  temperature,
  maxTokens,
  systemPrompt,
  onChange,
}: {
  nodeId: string
  provider: string
  model: string
  temperature?: number
  maxTokens?: number
  systemPrompt: string
  onChange: (
    nodeId: string,
    patch: {
      provider?: string
      model?: string
      temperature?: number
      maxTokens?: number
      systemPrompt?: string
    },
  ) => void
}) {
  return (
    <section className="space-y-3">
      <PanelSectionTitle title="LLM 配置" />

      <EditableTextField
        id={`llm-provider-${nodeId}`}
        label="Provider"
        value={provider}
        placeholder="例如：openai"
        onChange={(value) =>
          onChange(nodeId, {
            provider: value,
          })
        }
      />

      <EditableTextField
        id={`llm-model-${nodeId}`}
        label="Model"
        value={model}
        placeholder="例如：gpt-4.1-mini"
        onChange={(value) =>
          onChange(nodeId, {
            model: value,
          })
        }
      />

      <EditableNumberField
        id={`llm-temperature-${nodeId}`}
        label="Temperature"
        value={temperature}
        min={0}
        max={2}
        step={0.1}
        placeholder="0.7"
        onChange={(value) =>
          onChange(nodeId, {
            temperature: value,
          })
        }
      />

      <EditableNumberField
        id={`llm-max-tokens-${nodeId}`}
        label="Max Tokens"
        value={maxTokens}
        min={1}
        step={1}
        placeholder="1024"
        onChange={(value) =>
          onChange(nodeId, {
            maxTokens: value,
          })
        }
      />

      <EditableTextareaField
        id={`llm-system-prompt-${nodeId}`}
        label="System Prompt"
        value={systemPrompt}
        placeholder="例如：你是一个严谨的 AI 工作流助手。"
        rows={4}
        onChange={(value) =>
          onChange(nodeId, {
            systemPrompt: value,
          })
        }
      />
    </section>
  )
}

function ValidationSummaryCard({
  schema,
  validation,
}: {
  schema: WorkflowSchema
  validation: WorkflowValidationResult
}) {
  const hasErrors = validation.errorCount > 0
  const hasWarnings = !hasErrors && validation.warningCount > 0
  const statusLabel = hasErrors
    ? "存在错误"
    : hasWarnings
      ? "通过，但有警告"
      : "校验通过"
  const cardClassName = hasErrors
    ? "border-rose-200 bg-rose-50 text-rose-800"
    : hasWarnings
      ? "border-amber-200 bg-amber-50 text-amber-800"
      : "border-emerald-200 bg-emerald-50 text-emerald-800"

  return (
    <div className={`rounded-md border p-4 ${cardClassName}`}>
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0 space-y-1">
          <div className="flex items-center gap-2">
            {hasErrors ? (
              <AlertCircle className="h-4 w-4" />
            ) : hasWarnings ? (
              <TriangleAlert className="h-4 w-4" />
            ) : (
              <CheckCircle2 className="h-4 w-4" />
            )}
            <p className="text-sm font-semibold text-slate-950">{statusLabel}</p>
          </div>
          <p className="text-xs text-slate-600">
            当前结果基于画布中的 nodes / edges 实时计算。
          </p>
        </div>

        <Badge variant={hasErrors ? "destructive" : "secondary"}>
          {statusLabel}
        </Badge>
      </div>

      <div className="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
        <SummaryStat label="节点" value={schema.summary.node_count} />
        <SummaryStat label="连线" value={schema.summary.edge_count} />
        <SummaryStat label="Start" value={schema.summary.start_count} />
        <SummaryStat label="End" value={schema.summary.end_count} />
      </div>

      <div className="mt-3 flex flex-wrap gap-2">
        <Badge variant="outline">{validation.errorCount} 错误</Badge>
        <Badge variant="outline">{validation.warningCount} 警告</Badge>
      </div>
    </div>
  )
}

function ValidationIssueList({ issues }: { issues: WorkflowValidationIssue[] }) {
  if (issues.length === 0) {
    return (
      <div className="rounded-md border border-emerald-200 bg-emerald-50 p-4">
        <div className="flex items-start gap-3">
          <div className="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-emerald-100 text-emerald-700">
            <CheckCircle2 className="h-4 w-4" />
          </div>
          <div className="min-w-0">
            <p className="text-sm font-semibold text-emerald-900">没有发现问题</p>
            <p className="mt-1 text-xs text-emerald-800">当前 Workflow 的基础校验全部通过。</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {issues.map((issue) => (
        <ValidationIssueCard key={issue.id} issue={issue} />
      ))}
    </div>
  )
}

function ValidationIssueCard({ issue }: { issue: WorkflowValidationIssue }) {
  const isError = issue.severity === "error"

  return (
    <div
      className={`rounded-md border p-3 ${
        isError ? "border-rose-200 bg-rose-50" : "border-amber-200 bg-amber-50"
      }`}
    >
      <div className="flex items-start gap-3">
        <div
          className={`grid h-9 w-9 shrink-0 place-items-center rounded-md ${
            isError ? "bg-rose-100 text-rose-700" : "bg-amber-100 text-amber-700"
          }`}
        >
          {isError ? (
            <AlertCircle className="h-4 w-4" />
          ) : (
            <TriangleAlert className="h-4 w-4" />
          )}
        </div>

        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <p className="text-sm font-medium text-slate-950">{issue.message}</p>
            <Badge variant={isError ? "destructive" : "outline"} className="text-[10px]">
              {isError ? "错误" : "警告"}
            </Badge>
          </div>

          <p className="mt-1 text-xs text-slate-500">{issue.code}</p>

          {issue.nodeId || issue.edgeId ? (
            <div className="mt-2 flex flex-wrap gap-2">
              {issue.nodeId ? (
                <Badge variant="outline" className="text-[10px]">
                  node: {issue.nodeId}
                </Badge>
              ) : null}
              {issue.edgeId ? (
                <Badge variant="outline" className="text-[10px]">
                  edge: {issue.edgeId}
                </Badge>
              ) : null}
            </div>
          ) : null}
        </div>
      </div>
    </div>
  )
}

function EmptySelectionState() {
  return (
    <div className="flex min-h-[420px] items-center justify-center p-6">
      <div className="text-center">
        <div className="mx-auto grid h-10 w-10 place-items-center rounded-md bg-slate-100 text-slate-500">
          <SlidersHorizontal className="h-5 w-5" />
        </div>
        <p className="mt-3 text-sm font-medium text-slate-950">未选择对象</p>
        <p className="mt-1 text-xs text-slate-500">请选择节点、连线，或切换到校验 / Schema 标签页。</p>
      </div>
    </div>
  )
}

function PanelSectionTitle({ title }: { title: string }) {
  return (
    <div className="flex items-center gap-2">
      <Settings2 className="h-4 w-4 text-slate-500" />
      <h3 className="text-sm font-semibold text-slate-950">{title}</h3>
    </div>
  )
}

function EditableTextField({
  id,
  label,
  value,
  placeholder,
  onChange,
}: {
  id: string
  label: string
  value: string
  placeholder?: string
  onChange: (value: string) => void
}) {
  return (
    <div className="space-y-1.5">
      <Label htmlFor={id} className="text-xs font-medium text-slate-500">
        {label}
      </Label>
      <Input
        id={id}
        value={value}
        placeholder={placeholder}
        onChange={(event) => onChange(event.target.value)}
      />
    </div>
  )
}

function EditableNumberField({
  id,
  label,
  value,
  min,
  max,
  step,
  placeholder,
  onChange,
}: {
  id: string
  label: string
  value?: number
  min?: number
  max?: number
  step?: number
  placeholder?: string
  onChange: (value: number | undefined) => void
}) {
  return (
    <div className="space-y-1.5">
      <Label htmlFor={id} className="text-xs font-medium text-slate-500">
        {label}
      </Label>
      <Input
        id={id}
        type="number"
        value={value ?? ""}
        min={min}
        max={max}
        step={step}
        placeholder={placeholder}
        onChange={(event) => {
          const nextValue = event.target.value
          onChange(nextValue === "" ? undefined : Number(nextValue))
        }}
      />
    </div>
  )
}

function EditableTextareaField({
  id,
  label,
  value,
  placeholder,
  rows,
  onChange,
}: {
  id: string
  label: string
  value: string
  placeholder?: string
  rows?: number
  onChange: (value: string) => void
}) {
  return (
    <div className="space-y-1.5">
      <Label htmlFor={id} className="text-xs font-medium text-slate-500">
        {label}
      </Label>
      <Textarea
        id={id}
        value={value}
        placeholder={placeholder}
        rows={rows}
        onChange={(event) => onChange(event.target.value)}
      />
    </div>
  )
}

function ReadonlyPropertyRow({
  label,
  value,
}: {
  label: string
  value: string
}) {
  return (
    <div className="space-y-1">
      <dt className="text-xs font-medium text-slate-500">{label}</dt>
      <dd className="break-all rounded-md border border-slate-200 bg-slate-50 px-2 py-1.5 text-xs text-slate-800">
        {value}
      </dd>
    </div>
  )
}

function SummaryStat({
  label,
  value,
}: {
  label: string
  value: number
}) {
  return (
    <div className="rounded-md border border-slate-200 bg-white px-3 py-2">
      <div className="text-[11px] font-medium text-slate-500">{label}</div>
      <div className="mt-1 text-sm font-semibold text-slate-950">{value}</div>
    </div>
  )
}
