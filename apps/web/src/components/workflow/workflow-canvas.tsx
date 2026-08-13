"use client";

import {
  useCallback,
  useRef,
  useState,
  type DragEvent,
  type KeyboardEvent,
} from "react";
import {
  Background,
  Controls,
  MiniMap,
  ReactFlow,
  ReactFlowProvider,
  type NodeTypes,
  type ReactFlowInstance,
} from "reactflow";
import { RotateCcw, Trash2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  EndNode,
  GenericWorkflowNode,
  LLMNode,
  PromptNode,
  StartNode,
} from "@/components/workflow/nodes/basic-nodes";
import { useWorkflowDesignerStore } from "@/stores/workflow-designer-store";
import { useWorkflowStore } from "@/stores/workflow-store";
import {
  WORKFLOW_NODE_DRAG_TYPE,
  isWorkflowNodeType,
  type WorkflowDesignerEdge,
  type WorkflowDesignerNode,
} from "@/types/workflow";

const nodeTypes: NodeTypes = {
  start: StartNode,
  end: EndNode,
  prompt: PromptNode,
  llm: LLMNode,
  condition: GenericWorkflowNode,
  loop: GenericWorkflowNode,
  http: GenericWorkflowNode,
  tool: GenericWorkflowNode,
  memory: GenericWorkflowNode,
  rag: GenericWorkflowNode,
};

export function WorkflowCanvas() {
  return (
    <ReactFlowProvider>
      <WorkflowCanvasInner />
    </ReactFlowProvider>
  );
}

function WorkflowCanvasInner() {
  const flowWrapperRef = useRef<HTMLDivElement | null>(null);
  const [reactFlowInstance, setReactFlowInstance] =
    useState<ReactFlowInstance | null>(null);

  const nodes = useWorkflowDesignerStore((state) => state.nodes);
  const edges = useWorkflowDesignerStore((state) => state.edges);
  const selectedNodeId = useWorkflowDesignerStore((state) => state.selectedNodeId);
  const selectedEdgeId = useWorkflowDesignerStore((state) => state.selectedEdgeId);
  const draftName = useWorkflowStore((state) => state.draftName);
  const onNodesChange = useWorkflowDesignerStore((state) => state.onNodesChange);
  const onEdgesChange = useWorkflowDesignerStore((state) => state.onEdgesChange);
  const connectNodes = useWorkflowDesignerStore((state) => state.connectNodes);
  const addNodeFromTemplate = useWorkflowDesignerStore(
    (state) => state.addNodeFromTemplate,
  );
  const selectNode = useWorkflowDesignerStore((state) => state.selectNode);
  const selectEdge = useWorkflowDesignerStore((state) => state.selectEdge);
  const syncSelection = useWorkflowDesignerStore((state) => state.syncSelection);
  const clearSelection = useWorkflowDesignerStore((state) => state.clearSelection);
  const deleteSelection = useWorkflowDesignerStore(
    (state) => state.deleteSelection,
  );
  const resetDesigner = useWorkflowDesignerStore((state) => state.resetDesigner);

  const hasSelection = Boolean(selectedNodeId || selectedEdgeId);

  const handleDragOver = useCallback((event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  }, []);

  const handleDrop = useCallback(
    (event: DragEvent<HTMLDivElement>) => {
      event.preventDefault();

      if (!reactFlowInstance || !flowWrapperRef.current) {
        return;
      }

      const rawNodeType =
        event.dataTransfer.getData(WORKFLOW_NODE_DRAG_TYPE) ||
        event.dataTransfer.getData("text/plain");

      if (!isWorkflowNodeType(rawNodeType)) {
        return;
      }

      const bounds = flowWrapperRef.current.getBoundingClientRect();
      const position = reactFlowInstance.project({
        x: event.clientX - bounds.left,
        y: event.clientY - bounds.top,
      });

      addNodeFromTemplate(rawNodeType, {
        x: position.x - 104,
        y: position.y - 40,
      });
    },
    [addNodeFromTemplate, reactFlowInstance],
  );

  const handleKeyDown = useCallback(
    (event: KeyboardEvent<HTMLDivElement>) => {
      if (event.key !== "Delete" && event.key !== "Backspace") {
        return;
      }

      event.preventDefault();
      deleteSelection();
    },
    [deleteSelection],
  );

  const handleSelectionChange = useCallback(
    ({
      nodes: selectedNodes,
      edges: selectedEdges,
    }: {
      nodes: WorkflowDesignerNode[];
      edges: WorkflowDesignerEdge[];
    }) => {
      const selectedNode = selectedNodes[0] ?? null;
      const selectedEdge = selectedEdges[0] ?? null;

      syncSelection(selectedNode?.id ?? null, selectedEdge?.id ?? null);
    },
    [syncSelection],
  );

  return (
    <section className="flex min-h-[520px] min-w-0 flex-1 flex-col bg-slate-50">
      <div className="flex h-14 items-center justify-between border-b border-slate-200 bg-white px-4">
        <div className="min-w-0">
          <h2 className="truncate text-sm font-semibold text-slate-950">
            {draftName || "未命名 Workflow"}
          </h2>
          <p className="mt-0.5 text-xs text-slate-500">schema_version: 1.0</p>
        </div>

        <div className="flex items-center gap-2">
          <Badge variant="secondary">Draft</Badge>
          <Badge variant="outline">{nodes.length} 节点</Badge>
          <Badge variant="outline">{edges.length} 连线</Badge>

          {selectedNodeId ? (
            <Badge variant="outline">已选择节点</Badge>
          ) : selectedEdgeId ? (
            <Badge variant="outline">已选择连线</Badge>
          ) : null}

          <Button variant="outline" size="sm" onClick={resetDesigner}>
            <RotateCcw className="h-4 w-4" />
            重置
          </Button>

          <Button
            variant="destructive"
            size="sm"
            disabled={!hasSelection}
            onClick={deleteSelection}
          >
            <Trash2 className="h-4 w-4" />
            删除
          </Button>
        </div>
      </div>

      <div
        ref={flowWrapperRef}
        tabIndex={0}
        className="min-h-0 flex-1 outline-none"
        onDragOver={handleDragOver}
        onDrop={handleDrop}
        onKeyDown={handleKeyDown}
        onMouseDown={() => flowWrapperRef.current?.focus()}
      >
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          onInit={(instance) => setReactFlowInstance(instance)}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={connectNodes}
          onNodeClick={(_, node) => selectNode(node.id)}
          onEdgeClick={(_, edge) => selectEdge(edge.id)}
          onPaneClick={clearSelection}
          onSelectionChange={handleSelectionChange}
          fitView
          fitViewOptions={{
            padding: 0.25,
          }}
          nodesDraggable
          nodesConnectable
          elementsSelectable
          deleteKeyCode={null}
          proOptions={{
            hideAttribution: true,
          }}
        >
          <Background color="#cbd5e1" gap={24} size={1} />
          <Controls showInteractive={false} />
          <MiniMap pannable={false} zoomable={false} />
        </ReactFlow>
      </div>
    </section>
  );
}
