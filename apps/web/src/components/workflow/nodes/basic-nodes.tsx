"use client";

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
import { Handle, Position, type NodeProps } from "reactflow";

import { cn } from "@/lib/utils";
import type { WorkflowNodeData, WorkflowNodeType } from "@/types/workflow";

type BasicNodeProps = NodeProps<WorkflowNodeData>;

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

const nodeToneClassNameMap: Record<WorkflowNodeType, string> = {
  Start: "bg-emerald-50 text-emerald-700",
  End: "bg-rose-50 text-rose-700",
  Prompt: "bg-amber-50 text-amber-700",
  LLM: "bg-sky-50 text-sky-700",
  Condition: "bg-violet-50 text-violet-700",
  Loop: "bg-indigo-50 text-indigo-700",
  HTTP: "bg-cyan-50 text-cyan-700",
  Tool: "bg-orange-50 text-orange-700",
  Memory: "bg-teal-50 text-teal-700",
  RAG: "bg-lime-50 text-lime-700",
};

const handleClassNameMap: Record<WorkflowNodeType, string> = {
  Start: "!bg-emerald-500",
  End: "!bg-rose-500",
  Prompt: "!bg-amber-500",
  LLM: "!bg-sky-500",
  Condition: "!bg-violet-500",
  Loop: "!bg-indigo-500",
  HTTP: "!bg-cyan-500",
  Tool: "!bg-orange-500",
  Memory: "!bg-teal-500",
  RAG: "!bg-lime-500",
};

export function StartNode(props: BasicNodeProps) {
  return <WorkflowNodeFrame {...props} />;
}

export function EndNode(props: BasicNodeProps) {
  return <WorkflowNodeFrame {...props} />;
}

export function PromptNode(props: BasicNodeProps) {
  return <WorkflowNodeFrame {...props} />;
}

export function LLMNode(props: BasicNodeProps) {
  return <WorkflowNodeFrame {...props} />;
}

export function GenericWorkflowNode(props: BasicNodeProps) {
  return <WorkflowNodeFrame {...props} />;
}

function WorkflowNodeFrame({ data, selected }: BasicNodeProps) {
  const Icon = nodeIconMap[data.nodeType];

  return (
    <div className="relative">
      {hasTargetHandle(data.nodeType) ? (
        <Handle
          type="target"
          position={Position.Left}
          className={cn(
            "!h-3 !w-3 !border-2 !border-white",
            handleClassNameMap[data.nodeType],
          )}
        />
      ) : null}

      <div
        className={cn(
          "w-52 rounded-md border bg-white p-3 shadow-sm transition",
          selected
            ? "border-slate-950 ring-2 ring-slate-300"
            : "border-slate-200 hover:border-slate-300",
        )}
      >
        <div className="flex items-start gap-3">
          <div
            className={cn(
              "grid h-9 w-9 shrink-0 place-items-center rounded-md",
              nodeToneClassNameMap[data.nodeType],
            )}
          >
            <Icon className="h-4 w-4" />
          </div>

          <div className="min-w-0">
            <div className="truncate text-sm font-semibold text-slate-950">
              {data.label}
            </div>
            <div className="mt-1 text-xs text-slate-500">
              {data.description}
            </div>
          </div>
        </div>
      </div>

      {hasSourceHandle(data.nodeType) ? (
        <Handle
          type="source"
          position={Position.Right}
          className={cn(
            "!h-3 !w-3 !border-2 !border-white",
            handleClassNameMap[data.nodeType],
          )}
        />
      ) : null}
    </div>
  );
}

function hasTargetHandle(nodeType: WorkflowNodeType) {
  return nodeType !== "Start";
}

function hasSourceHandle(nodeType: WorkflowNodeType) {
  return nodeType !== "End";
}