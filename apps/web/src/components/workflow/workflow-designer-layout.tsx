"use client";

import { NodePalette } from "@/components/workflow/node-palette";
import { PropertiesPanel } from "@/components/workflow/properties-panel";
import { WorkflowCanvas } from "@/components/workflow/workflow-canvas";

export function WorkflowDesignerLayout() {
  return (
    <section className="flex min-h-[720px] flex-col overflow-hidden rounded-lg border border-slate-200 bg-white shadow-sm xl:h-[calc(100vh-10rem)] xl:flex-row">
      <NodePalette />
      <WorkflowCanvas />
      <PropertiesPanel />
    </section>
  );
}