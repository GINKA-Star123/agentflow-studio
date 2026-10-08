"use client";

import type { DragEvent } from "react";
import type { LucideIcon } from "lucide-react";
import {
  Bot,
  Database,
  GitBranch,
  Globe,
  MessageSquare,
  Play,
  Repeat,
  Search,
  Square,
  Wrench,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  WORKFLOW_NODE_DRAG_TYPE,
  workflowNodeGroups,
  workflowNodeTemplates,
  type WorkflowNodeTemplate,
  type WorkflowNodeType,
} from "@/types/workflow";

const nodeIconMap: Record<WorkflowNodeType, LucideIcon> = {
  Start: Play,
  End: Square,
  Prompt: MessageSquare,
  LLM: Bot,
  Condition: GitBranch,
  Loop: Repeat,
  HTTP: Globe,
  Tool: Wrench,
  Memory: Database,
  RAG: Search,
};

const groupedTemplates = workflowNodeGroups
  .map((group) => ({
    group,
    templates: workflowNodeTemplates.filter((template) => template.group === group),
  }))
  .filter((item) => item.templates.length > 0);

export function NodePalette() {
  return (
    <aside className="flex min-h-0 w-full flex-col border-b border-slate-200 bg-white xl:w-72 xl:border-b-0 xl:border-r">
      <div className="border-b border-slate-200 px-4 py-4">
        <h2 className="text-sm font-semibold text-slate-950">节点面板</h2>
        <p className="mt-1 text-xs text-slate-500">Workflow 节点类型</p>
      </div>

      <ScrollArea className="flex-1">
        <div className="space-y-5 p-4">
          {groupedTemplates.map(({ group, templates }) => (
            <section key={group} className="space-y-2">
              <div className="flex items-center justify-between">
                <h3 className="text-xs font-medium text-slate-500">{group}</h3>
                <Badge variant="outline" className="text-[10px]">
                  {templates.length}
                </Badge>
              </div>

              <div className="space-y-2">
                {templates.map((template) => (
                  <PaletteNodeCard key={template.type} template={template} />
                ))}
              </div>
            </section>
          ))}
        </div>
      </ScrollArea>
    </aside>
  );
}

function PaletteNodeCard({ template }: { template: WorkflowNodeTemplate }) {
  const Icon = nodeIconMap[template.type];

  function handleDragStart(event: DragEvent<HTMLButtonElement>) {
    event.dataTransfer.setData(WORKFLOW_NODE_DRAG_TYPE, template.type);
    event.dataTransfer.setData("text/plain", template.type);
    event.dataTransfer.effectAllowed = "move";
  }

  return (
    <button
      type="button"
      draggable
      onDragStart={handleDragStart}
      className="w-full cursor-grab rounded-md border border-slate-200 bg-white p-3 text-left shadow-sm transition hover:border-slate-300 hover:bg-slate-50 active:cursor-grabbing"
    >
      <div className="flex items-start gap-3">
        <div className="grid h-9 w-9 shrink-0 place-items-center rounded-md bg-slate-100 text-slate-700">
          <Icon className="h-4 w-4" />
        </div>

        <div className="min-w-0">
          <div className="truncate text-sm font-medium text-slate-950">
            {template.label}
          </div>
          <div className="mt-1 text-xs text-slate-500">
            {template.description}
          </div>
        </div>
      </div>
    </button>
  );
}