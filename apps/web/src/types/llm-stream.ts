export type LLMStreamEventType = "start" | "delta" | "usage" | "done" | "error"

export type LLMStreamTokenUsage = {
  input_tokens?: number
  output_tokens?: number
  total_tokens?: number
}

export type LLMStreamError = {
  code: string
  message: string
  retryable?: boolean
  details?: unknown
}

export type LLMStreamEvent = {
  type: LLMStreamEventType
  delta?: string
  text?: string
  finish_reason?: string
  token_usage?: LLMStreamTokenUsage | null
  error?: LLMStreamError | null
  metadata?: Record<string, unknown>
  request_id?: string
  gateway_request_id?: string
}

export type LLMStreamMessage = {
  role: "system" | "user" | "assistant" | "tool"
  content: string
}

export type LLMStreamRequest = {
  provider: string
  model: string
  messages: LLMStreamMessage[]
  temperature?: number
  max_tokens?: number
  top_p?: number
  frequency_penalty?: number
  presence_penalty?: number
  stop?: string | string[]
  metadata?: Record<string, unknown>
}