import { create } from "zustand"

import { workflowApi } from "@/lib/api-client"
import { useWorkflowDesignerStore } from "@/stores/workflow-designer-store"
import type {
  WorkflowDefinition,
  WorkflowDefinitionSummary,
  WorkflowListResult,
} from "@/types/workflow-definition"

type WorkflowRequestStatus = "idle" | "loading" | "saving" | "error"

type WorkflowState = {
  status: WorkflowRequestStatus
  currentWorkspaceId: string | null
  currentWorkflowId: string | null
  currentWorkflow: WorkflowDefinition | null
  draftName: string
  items: WorkflowDefinitionSummary[]
  errorMessage: string

  setDraftName: (name: string) => void
  loadList: (workspaceId: string) => Promise<WorkflowListResult | null>
  loadWorkflow: (input: {
    workspaceId: string
    workflowId: string
  }) => Promise<WorkflowDefinition | null>
  saveWorkflow: (input: {
    workspaceId: string
    workflowId?: string | null
    name: string
    schema: WorkflowDefinition["schema"]
  }) => Promise<WorkflowDefinition | null>
  clearCurrent: () => void
}

export const useWorkflowStore = create<WorkflowState>((set) => ({
  status: "idle",
  currentWorkspaceId: null,
  currentWorkflowId: null,
  currentWorkflow: null,
  draftName: "未命名 Workflow",
  items: [],
  errorMessage: "",

  setDraftName(name) {
    set({ draftName: name, errorMessage: "" })
  },

  async loadList(workspaceId) {
    set({ status: "loading", errorMessage: "" })

    try {
      const result = await workflowApi.list(workspaceId)
      set({ status: "idle", items: result.items, errorMessage: "" })
      return result
    } catch (error) {
      set({ status: "error", errorMessage: getErrorMessage(error) })
      return null
    }
  },

  async loadWorkflow({ workspaceId, workflowId }) {
    set({ status: "loading", errorMessage: "" })

    try {
      const result = await workflowApi.get(workspaceId, workflowId)
      useWorkflowDesignerStore.getState().loadFromSchema(result.schema)
      set({
        status: "idle",
        currentWorkspaceId: workspaceId,
        currentWorkflowId: result.id,
        currentWorkflow: result,
        draftName: result.name,
        errorMessage: "",
      })
      return result
    } catch (error) {
      set({ status: "error", errorMessage: getErrorMessage(error) })
      return null
    }
  },

  async saveWorkflow({ workspaceId, workflowId, name, schema }) {
    set({ status: "saving", errorMessage: "" })

    try {
      const result = workflowId
        ? await workflowApi.update(workspaceId, workflowId, { name, schema })
        : await workflowApi.create(workspaceId, { name, schema })

      set({
        status: "idle",
        currentWorkspaceId: workspaceId,
        currentWorkflowId: result.id,
        currentWorkflow: result,
        draftName: result.name,
        items: [
          result,
          ...useWorkflowStore
            .getState()
            .items.filter((item) => item.id !== result.id),
        ],
        errorMessage: "",
      })
      return result
    } catch (error) {
      set({ status: "error", errorMessage: getErrorMessage(error) })
      return null
    }
  },

  clearCurrent() {
    set({
      status: "idle",
      currentWorkspaceId: null,
      currentWorkflowId: null,
      currentWorkflow: null,
      draftName: "未命名 Workflow",
      errorMessage: "",
    })
  },
}))

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : "Workflow 请求失败"
}
