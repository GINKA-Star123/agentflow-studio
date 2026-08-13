"use client"

import { NodePalette } from "@/components/workflow/node-palette"
import { PropertiesPanel } from "@/components/workflow/properties-panel"
import { WorkflowCanvas } from "@/components/workflow/workflow-canvas"
import { WorkflowRunPanel } from "@/components/workflow/workflow-run-panel"
import { WorkflowRunToolbar } from "@/components/workflow/workflow-run-toolbar"
import { WorkflowStreamPanel } from "@/components/workflow/workflow-stream-panel"

export function WorkflowDesignerLayout() {
  return (
    <section className="flex min-h-[720px] flex-col overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm xl:h-[calc(100vh-10rem)] xl:flex-row">
      <NodePalette />

      <div className="flex min-h-0 min-w-0 flex-1 flex-col">
        <WorkflowRunToolbar />
        <WorkflowCanvas />

        <div className="grid border-t border-slate-200 xl:grid-cols-2">
          <WorkflowRunPanel />
          <WorkflowStreamPanel />
        </div>
      </div>

      <PropertiesPanel />
    </section>
  )
}