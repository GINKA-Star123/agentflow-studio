import { create } from "zustand"

import { workflowRunApi } from "@/lib/api-client"
import type {
  StartWorkflowRunPayload,
  WorkflowRunDetail,
  WorkflowRunNodeExecution,
} from "@/types/workflow-run"

type WorkflowRunRequestStatus =
  | "idle"
  | "running"
  | "refreshing"
  | "canceling"
  | "failed"

type StartRunInput = {
  workspaceId: string
  workflowId: string
  input: Record<string, unknown>
  traceId?: string
}

type RefreshRunInput = {
  workspaceId: string
  runId?: string
}

type CancelRunInput = {
  workspaceId: string
  runId?: string
}

type WorkflowRunState = {
  requestStatus: WorkflowRunRequestStatus
  currentRun: WorkflowRunDetail | null
  nodeExecutions: WorkflowRunNodeExecution[]
  errorMessage: string

  startRun: (input: StartRunInput) => Promise<void>
  refreshRun: (input: RefreshRunInput) => Promise<void>
  cancelRun: (input: CancelRunInput) => Promise<void>
  setErrorMessage: (message: string) => void
  resetRunState: () => void
}

export const useWorkflowRunStore = create<WorkflowRunState>((set, get) => ({
  requestStatus: "idle",
  currentRun: null,
  nodeExecutions: [],
  errorMessage: "",

  async startRun(input) {
    set({
      requestStatus: "running",
      errorMessage: "",
      currentRun: null,
      nodeExecutions: [],
    })

    try {
      const payload: StartWorkflowRunPayload = {
        input: input.input,
        trace_id: input.traceId,
      }

      const run = await workflowRunApi.start(
        input.workspaceId,
        input.workflowId,
        payload,
      )

      const nodesResult = await workflowRunApi.listNodes(
        input.workspaceId,
        run.id,
      )

      set({
        requestStatus: "idle",
        currentRun: run,
        nodeExecutions: nodesResult.items,
        errorMessage: "",
      })
    } catch (error) {
      set({
        requestStatus: "failed",
        errorMessage: getErrorMessage(error),
      })
    }
  },

  async refreshRun(input) {
    const runId = input.runId ?? get().currentRun?.id
    if (!runId) {
      set({
        requestStatus: "failed",
        errorMessage: "暂无可刷新的 Run",
      })
      return
    }

    set({
      requestStatus: "refreshing",
      errorMessage: "",
    })

    try {
      const [run, nodesResult] = await Promise.all([
        workflowRunApi.get(input.workspaceId, runId),
        workflowRunApi.listNodes(input.workspaceId, runId),
      ])

      set({
        requestStatus: "idle",
        currentRun: run,
        nodeExecutions: nodesResult.items,
        errorMessage: "",
      })
    } catch (error) {
      set({
        requestStatus: "failed",
        errorMessage: getErrorMessage(error),
      })
    }
  },

  async cancelRun(input) {
    const runId = input.runId ?? get().currentRun?.id
    if (!runId) {
      set({
        requestStatus: "failed",
        errorMessage: "暂无可取消的 Run",
      })
      return
    }

    set({
      requestStatus: "canceling",
      errorMessage: "",
    })

    try {
      const result = await workflowRunApi.cancel(input.workspaceId, runId)
      const nodesResult = await workflowRunApi.listNodes(
        input.workspaceId,
        result.run.id,
      )

      set({
        requestStatus: "idle",
        currentRun: result.run,
        nodeExecutions: nodesResult.items,
        errorMessage: "",
      })
    } catch (error) {
      set({
        requestStatus: "failed",
        errorMessage: getErrorMessage(error),
      })
    }
  },

  setErrorMessage(message) {
    set({
      requestStatus: "failed",
      errorMessage: message,
    })
  },

  resetRunState() {
    set({
      requestStatus: "idle",
      currentRun: null,
      nodeExecutions: [],
      errorMessage: "",
    })
  },
}))

function getErrorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message
  }

  return "Workflow Run 请求失败"
}
