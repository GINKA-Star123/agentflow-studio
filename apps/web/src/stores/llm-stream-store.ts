import { create } from "zustand"

import { streamLLM } from "@/lib/llm-stream-client"
import type {
  LLMStreamEvent,
  LLMStreamRequest,
  LLMStreamTokenUsage,
} from "@/types/llm-stream"

type LLMStreamStatus = "idle" | "connecting" | "streaming" | "done" | "failed" | "canceled"

type StartLLMStreamInput = {
  workspaceId: string
  payload: LLMStreamRequest
}

type LLMStreamState = {
  status: LLMStreamStatus
  events: LLMStreamEvent[]
  responseText: string
  tokenUsage: LLMStreamTokenUsage | null
  errorMessage: string
  requestId: string
  gatewayRequestId: string
  startedAt: number | null
  finishedAt: number | null
  abortController: AbortController | null

  startStream: (input: StartLLMStreamInput) => Promise<void>
  cancelStream: () => void
  clearStream: () => void
}

export const useLLMStreamStore = create<LLMStreamState>((set, get) => ({
  status: "idle",
  events: [],
  responseText: "",
  tokenUsage: null,
  errorMessage: "",
  requestId: "",
  gatewayRequestId: "",
  startedAt: null,
  finishedAt: null,
  abortController: null,

  async startStream(input) {
    const currentController = get().abortController
    currentController?.abort()

    const abortController = new AbortController()

    set({
      status: "connecting",
      events: [],
      responseText: "",
      tokenUsage: null,
      errorMessage: "",
      requestId: "",
      gatewayRequestId: "",
      startedAt: Date.now(),
      finishedAt: null,
      abortController,
    })

    try {
      await streamLLM({
        workspaceId: input.workspaceId,
        payload: input.payload,
        signal: abortController.signal,
        onEvent(event) {
          appendStreamEvent(event, set)
        },
      })

      const currentStatus = get().status
      if (currentStatus !== "failed" && currentStatus !== "canceled") {
        set({
          status: "done",
          finishedAt: Date.now(),
          abortController: null,
        })
      }
    } catch (error) {
      if (abortController.signal.aborted) {
        set({
          status: "canceled",
          finishedAt: Date.now(),
          abortController: null,
        })
        return
      }

      set({
        status: "failed",
        errorMessage: getErrorMessage(error),
        finishedAt: Date.now(),
        abortController: null,
      })
    }
  },

  cancelStream() {
    const abortController = get().abortController
    abortController?.abort()

    set({
      status: "canceled",
      finishedAt: Date.now(),
      abortController: null,
    })
  },

  clearStream() {
    const abortController = get().abortController
    abortController?.abort()

    set({
      status: "idle",
      events: [],
      responseText: "",
      tokenUsage: null,
      errorMessage: "",
      requestId: "",
      gatewayRequestId: "",
      startedAt: null,
      finishedAt: null,
      abortController: null,
    })
  },
}))

function appendStreamEvent(
  event: LLMStreamEvent,
  set: (
    updater: (state: LLMStreamState) => Partial<LLMStreamState>,
  ) => void,
) {
  set((state) => {
    const nextEvents = [...state.events, event]

    const nextState: Partial<LLMStreamState> = {
      events: nextEvents,
      requestId: event.request_id ?? state.requestId,
      gatewayRequestId: event.gateway_request_id ?? state.gatewayRequestId,
    }

    if (event.type === "start") {
      nextState.status = "streaming"
    }

    if (event.type === "delta") {
      nextState.status = "streaming"
      nextState.responseText = event.text ?? `${state.responseText}${event.delta ?? ""}`
    }

    if (event.type === "usage" && event.token_usage) {
      nextState.tokenUsage = event.token_usage
    }

    if (event.type === "done") {
      nextState.status = "done"
      nextState.responseText = event.text ?? state.responseText
      nextState.tokenUsage = event.token_usage ?? state.tokenUsage
      nextState.finishedAt = Date.now()
    }

    if (event.type === "error") {
      nextState.status = "failed"
      nextState.errorMessage = event.error?.message ?? "流式输出失败"
      nextState.finishedAt = Date.now()
    }

    return nextState
  })
}

function getErrorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message
  }

  return "流式请求失败"
}
