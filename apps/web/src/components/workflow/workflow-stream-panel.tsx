"use client"

import {
  Activity,
  AlertCircle,
  CheckCircle2,
  Clock3,
  Radio,
  Trash2,
} from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import { useLLMStreamStore } from "@/stores/llm-stream-store"
import type { LLMStreamEvent, LLMStreamTokenUsage } from "@/types/llm-stream"

export function WorkflowStreamPanel() {
  const status = useLLMStreamStore((state) => state.status)
  const events = useLLMStreamStore((state) => state.events)
  const responseText = useLLMStreamStore((state) => state.responseText)
  const tokenUsage = useLLMStreamStore((state) => state.tokenUsage)
  const errorMessage = useLLMStreamStore((state) => state.errorMessage)
  const requestId = useLLMStreamStore((state) => state.requestId)
  const gatewayRequestId = useLLMStreamStore((state) => state.gatewayRequestId)
  const startedAt = useLLMStreamStore((state) => state.startedAt)
  const finishedAt = useLLMStreamStore((state) => state.finishedAt)
  const clearStream = useLLMStreamStore((state) => state.clearStream)

  const latencyMs = calculateLatencyMs(startedAt, finishedAt)
  const hasOutput = Boolean(responseText || events.length || errorMessage)

  return (
    <section className="min-h-72 border-t border-slate-200 bg-white">
      <div className="flex h-12 items-center justify-between border-b border-slate-200 px-4">
        <div className="flex min-w-0 items-center gap-2">
          <Radio className="h-4 w-4 text-slate-500" />
          <h2 className="truncate text-sm font-semibold text-slate-950">
            流式输出
          </h2>
        </div>

        <div className="flex shrink-0 items-center gap-2">
          <Badge variant={getStatusBadgeVariant(status)}>{status}</Badge>

          <Button
            variant="outline"
            size="sm"
            disabled={!hasOutput}
            onClick={clearStream}
          >
            <Trash2 className="h-4 w-4" />
            清空
          </Button>
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

          {!hasOutput ? <EmptyStreamState /> : null}

          {hasOutput ? (
            <StreamSummary
              eventCount={events.length}
              latencyMs={latencyMs}
              requestId={requestId}
              gatewayRequestId={gatewayRequestId}
            />
          ) : null}

          {responseText ? (
            <section className="space-y-2">
              <h3 className="text-xs font-semibold uppercase text-slate-500">
                Response Text
              </h3>

              <div className="rounded-md border border-slate-200 bg-slate-50 p-3">
                <p className="whitespace-pre-wrap text-sm leading-6 text-slate-800">
                  {responseText}
                </p>
              </div>
            </section>
          ) : null}

          {tokenUsage ? <TokenUsageBar tokenUsage={tokenUsage} /> : null}

          {events.length > 0 ? (
            <section className="space-y-2">
              <h3 className="text-xs font-semibold uppercase text-slate-500">
                Stream Events
              </h3>

              <div className="space-y-2">
                {events.map((event, index) => (
                  <StreamEventRow
                    key={`${event.type}-${index}`}
                    event={event}
                    index={index}
                  />
                ))}
              </div>
            </section>
          ) : null}
        </div>
      </ScrollArea>
    </section>
  )
}

function EmptyStreamState() {
  return (
    <div className="flex min-h-36 items-center justify-center rounded-md border border-dashed border-slate-200 bg-slate-50">
      <div className="text-center">
        <Clock3 className="mx-auto h-5 w-5 text-slate-400" />
        <p className="mt-2 text-sm font-medium text-slate-700">
          暂无流式输出
        </p>
      </div>
    </div>
  )
}

function StreamSummary({
  eventCount,
  latencyMs,
  requestId,
  gatewayRequestId,
}: {
  eventCount: number
  latencyMs: number | null
  requestId: string
  gatewayRequestId: string
}) {
  return (
    <div className="rounded-md border border-slate-200 bg-slate-50 p-3">
      <div className="flex flex-wrap gap-2">
        <Badge variant="outline">{eventCount} events</Badge>

        {latencyMs !== null ? (
          <Badge variant="outline">{latencyMs} ms</Badge>
        ) : null}

        {requestId ? (
          <Badge variant="outline" className="max-w-full truncate">
            runtime: {requestId}
          </Badge>
        ) : null}

        {gatewayRequestId ? (
          <Badge variant="outline" className="max-w-full truncate">
            gateway: {gatewayRequestId}
          </Badge>
        ) : null}
      </div>
    </div>
  )
}

function TokenUsageBar({
  tokenUsage,
}: {
  tokenUsage: LLMStreamTokenUsage
}) {
  return (
    <section className="space-y-2">
      <h3 className="text-xs font-semibold uppercase text-slate-500">
        Token Usage
      </h3>

      <div className="flex flex-wrap gap-2 rounded-md border border-slate-200 bg-white p-3">
        <Badge variant="outline">input {tokenUsage.input_tokens ?? 0}</Badge>
        <Badge variant="outline">output {tokenUsage.output_tokens ?? 0}</Badge>
        <Badge variant="outline">total {tokenUsage.total_tokens ?? 0}</Badge>
      </div>
    </section>
  )
}

function StreamEventRow({
  event,
  index,
}: {
  event: LLMStreamEvent
  index: number
}) {
  return (
    <div className="rounded-md border border-slate-200 bg-white p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex min-w-0 items-center gap-2">
          {event.type === "done" ? (
            <CheckCircle2 className="h-4 w-4 shrink-0 text-emerald-600" />
          ) : event.type === "error" ? (
            <AlertCircle className="h-4 w-4 shrink-0 text-rose-600" />
          ) : (
            <Activity className="h-4 w-4 shrink-0 text-slate-500" />
          )}

          <p className="text-sm font-medium text-slate-950">#{index + 1}</p>

          <Badge variant={getEventBadgeVariant(event.type)}>{event.type}</Badge>
        </div>

        {event.finish_reason ? (
          <Badge variant="outline">{event.finish_reason}</Badge>
        ) : null}
      </div>

      {event.delta ? (
        <p className="mt-2 whitespace-pre-wrap text-sm leading-6 text-slate-700">
          {event.delta}
        </p>
      ) : null}

      {event.error ? (
        <pre className="mt-2 max-h-28 overflow-auto rounded-md bg-rose-950 p-3 text-xs leading-5 text-rose-50">
          {JSON.stringify(event.error, null, 2)}
        </pre>
      ) : null}
    </div>
  )
}

function getStatusBadgeVariant(status: string) {
  switch (status) {
    case "done":
      return "secondary"

    case "failed":
    case "canceled":
      return "destructive"

    default:
      return "outline"
  }
}

function getEventBadgeVariant(type: string) {
  switch (type) {
    case "done":
    case "usage":
      return "secondary"

    case "error":
      return "destructive"

    default:
      return "outline"
  }
}

function calculateLatencyMs(startedAt: number | null, finishedAt: number | null) {
  if (!startedAt) {
    return null
  }

  const end = finishedAt ?? Date.now()
  return Math.max(0, end - startedAt)
}