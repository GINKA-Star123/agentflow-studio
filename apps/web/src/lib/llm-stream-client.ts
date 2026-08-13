import { getAccessToken } from "@/lib/auth-token"
import type { LLMStreamEvent, LLMStreamRequest } from "@/types/llm-stream"

const API_BASE_PATH = process.env.NEXT_PUBLIC_API_BASE_PATH ?? "/api/v1"

export type StreamLLMOptions = {
  workspaceId: string
  payload: LLMStreamRequest
  signal?: AbortSignal
  onEvent: (event: LLMStreamEvent) => void
}

export async function streamLLM(options: StreamLLMOptions): Promise<void> {
  const token = getAccessToken()
  const headers = new Headers()

  headers.set("Accept", "text/event-stream")
  headers.set("Content-Type", "application/json")

  if (token) {
    headers.set("Authorization", `Bearer ${token}`)
  }

  const response = await fetch(
    `${API_BASE_PATH}/workspaces/${options.workspaceId}/llm/stream`,
    {
      method: "POST",
      headers,
      body: JSON.stringify(options.payload),
      cache: "no-store",
      signal: options.signal,
    },
  )

  if (!response.ok) {
    throw await createStreamHTTPError(response)
  }

  if (!response.body) {
    throw new Error("流式响应体为空")
  }

  await readSSEStream(response.body, options.onEvent)
}

async function readSSEStream(
  body: ReadableStream<Uint8Array>,
  onEvent: (event: LLMStreamEvent) => void,
) {
  const reader = body.getReader()
  const decoder = new TextDecoder("utf-8")

  let buffer = ""

  while (true) {
    const { value, done } = await reader.read()

    if (done) {
      break
    }

    buffer += decoder.decode(value, {
      stream: true,
    })

    const events = splitSSEBuffer(buffer)
    buffer = events.rest

    for (const rawEvent of events.items) {
      const event = parseSSEEvent(rawEvent)
      if (event) {
        onEvent(event)
      }
    }
  }

  buffer += decoder.decode()

  const trailingEvent = parseSSEEvent(buffer)
  if (trailingEvent) {
    onEvent(trailingEvent)
  }
}

function splitSSEBuffer(buffer: string) {
  const normalized = buffer.replaceAll("\r\n", "\n")
  const parts = normalized.split("\n\n")
  const rest = parts.pop() ?? ""

  return {
    items: parts,
    rest,
  }
}

function parseSSEEvent(rawEvent: string): LLMStreamEvent | null {
  const lines = rawEvent.split("\n")

  let eventType = ""
  const dataLines: string[] = []

  for (const line of lines) {
    if (!line || line.startsWith(":")) {
      continue
    }

    const separatorIndex = line.indexOf(":")
    if (separatorIndex < 0) {
      continue
    }

    const field = line.slice(0, separatorIndex)
    const value = line.slice(separatorIndex + 1).replace(/^ /, "")

    if (field === "event") {
      eventType = value.trim()
    }

    if (field === "data") {
      dataLines.push(value)
    }
  }

  const data = dataLines.join("\n").trim()
  if (!data) {
    return null
  }

  if (data === "[DONE]") {
    return {
      type: "done",
      finish_reason: "stop",
    }
  }

  const parsed = JSON.parse(data) as LLMStreamEvent

  if (!parsed.type && eventType) {
    parsed.type = eventType as LLMStreamEvent["type"]
  }

  return parsed
}

async function createStreamHTTPError(response: Response) {
  const contentType = response.headers.get("content-type") ?? ""

  if (contentType.includes("application/json")) {
    try {
      const payload = (await response.json()) as {
        error?: {
          code?: string
          message?: string
          details?: unknown
        }
        request_id?: string
      }

      return new Error(
        payload.error?.message ??
          `流式请求失败：${response.status}`,
      )
    } catch {
      return new Error(`流式请求失败：${response.status}`)
    }
  }

  const text = await response.text()
  return new Error(text || `流式请求失败：${response.status}`)
}