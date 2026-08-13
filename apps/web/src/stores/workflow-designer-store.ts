import {
  MarkerType,
  applyEdgeChanges,
  applyNodeChanges,
  type Connection,
  type EdgeChange,
  type NodeChange,
  type XYPosition,
} from "reactflow";
import { create } from "zustand";

import {
  createDefaultWorkflowNodeConfig,
  getWorkflowNodeRenderType,
  getWorkflowNodeTemplate,
  type WorkflowDesignerEdge,
  type WorkflowDesignerNode,
  type WorkflowNodeConfig,
  type WorkflowNodeData,
  type WorkflowNodeTemplate,
  type WorkflowNodeType,
} from "@/types/workflow";
import type { WorkflowSchema } from "@/lib/workflow-schema";

type WorkflowDesignerState = {
  nodes: WorkflowDesignerNode[];
  edges: WorkflowDesignerEdge[];
  selectedNodeId: string | null;
  selectedEdgeId: string | null;

  onNodesChange: (changes: NodeChange[]) => void;
  onEdgesChange: (changes: EdgeChange[]) => void;
  addNodeFromTemplate: (nodeType: WorkflowNodeType, position: XYPosition) => void;
  connectNodes: (connection: Connection) => void;
  selectNode: (nodeId: string) => void;
  selectEdge: (edgeId: string) => void;
  syncSelection: (nodeId: string | null, edgeId: string | null) => void;
  clearSelection: () => void;
  deleteSelection: () => void;
  deleteNode: (nodeId: string) => void;
  deleteEdge: (edgeId: string) => void;
  updateNodeData: (
    nodeId: string,
    patch: Partial<Pick<WorkflowNodeData, "label" | "description">>,
  ) => void;
  updateNodeConfig: (nodeId: string, patch: Partial<WorkflowNodeConfig>) => void;
  loadFromSchema: (schema: WorkflowSchema) => void;
  resetDesigner: () => void;
};

export const useWorkflowDesignerStore = create<WorkflowDesignerState>(
  (set, get) => ({
    nodes: createInitialNodes(),
    edges: createInitialEdges(),
    selectedNodeId: null,
    selectedEdgeId: null,

    onNodesChange(changes) {
      set((state) => {
        const removedNodeIds = new Set(
          changes
            .filter((change) => change.type === "remove")
            .map((change) => change.id),
        );

        const nextNodes = applyNodeChanges(changes, state.nodes);
        const nextEdges =
          removedNodeIds.size > 0
            ? state.edges.filter(
                (edge) =>
                  !removedNodeIds.has(edge.source) &&
                  !removedNodeIds.has(edge.target),
              )
            : state.edges;

        return {
          nodes: nextNodes,
          edges: nextEdges,
          selectedNodeId:
            state.selectedNodeId && removedNodeIds.has(state.selectedNodeId)
              ? null
              : state.selectedNodeId,
          selectedEdgeId:
            state.selectedEdgeId &&
            !nextEdges.some((edge) => edge.id === state.selectedEdgeId)
              ? null
              : state.selectedEdgeId,
        };
      });
    },

    onEdgesChange(changes) {
      set((state) => {
        const removedEdgeIds = new Set(
          changes
            .filter((change) => change.type === "remove")
            .map((change) => change.id),
        );

        return {
          edges: applyEdgeChanges(changes, state.edges),
          selectedEdgeId:
            state.selectedEdgeId && removedEdgeIds.has(state.selectedEdgeId)
              ? null
              : state.selectedEdgeId,
        };
      });
    },

    addNodeFromTemplate(nodeType, position) {
      const template = getWorkflowNodeTemplate(nodeType);
      const nodeId = createNodeId(nodeType);

      const nextNode: WorkflowDesignerNode = {
        ...createWorkflowNode(template, nodeId, position),
        selected: true,
      };

      set((state) => ({
        nodes: [
          ...state.nodes.map((node) => ({ ...node, selected: false })),
          nextNode,
        ],
        edges: state.edges.map((edge) => ({ ...edge, selected: false })),
        selectedNodeId: nodeId,
        selectedEdgeId: null,
      }));
    },

    connectNodes(connection) {
      const { source, sourceHandle, target, targetHandle } = connection;

      if (!source || !target || source === target) {
        return;
      }

      set((state) => {
        const existingEdge = state.edges.find(
          (edge) =>
            edge.source === source &&
            edge.target === target &&
            edge.sourceHandle === sourceHandle &&
            edge.targetHandle === targetHandle,
        );

        if (existingEdge) {
          return {
            nodes: state.nodes.map((node) => ({ ...node, selected: false })),
            edges: state.edges.map((edge) => ({
              ...edge,
              selected: edge.id === existingEdge.id,
            })),
            selectedNodeId: null,
            selectedEdgeId: existingEdge.id,
          };
        }

        const nextEdge = createWorkflowEdge({
          source,
          sourceHandle,
          target,
          targetHandle,
          selected: true,
        });

        return {
          nodes: state.nodes.map((node) => ({ ...node, selected: false })),
          edges: [
            ...state.edges.map((edge) => ({ ...edge, selected: false })),
            nextEdge,
          ],
          selectedNodeId: null,
          selectedEdgeId: nextEdge.id,
        };
      });
    },

    selectNode(nodeId) {
      set((state) => ({
        nodes: state.nodes.map((node) => ({
          ...node,
          selected: node.id === nodeId,
        })),
        edges: state.edges.map((edge) => ({ ...edge, selected: false })),
        selectedNodeId: nodeId,
        selectedEdgeId: null,
      }));
    },

    selectEdge(edgeId) {
      set((state) => ({
        nodes: state.nodes.map((node) => ({ ...node, selected: false })),
        edges: state.edges.map((edge) => ({
          ...edge,
          selected: edge.id === edgeId,
        })),
        selectedNodeId: null,
        selectedEdgeId: edgeId,
      }));
    },

    syncSelection(nodeId, edgeId) {
      set({
        selectedNodeId: nodeId,
        selectedEdgeId: nodeId ? null : edgeId,
      });
    },

    clearSelection() {
      set((state) => ({
        nodes: state.nodes.map((node) => ({ ...node, selected: false })),
        edges: state.edges.map((edge) => ({ ...edge, selected: false })),
        selectedNodeId: null,
        selectedEdgeId: null,
      }));
    },

    deleteSelection() {
      const { selectedNodeId, selectedEdgeId, deleteNode, deleteEdge } = get();

      if (selectedNodeId) {
        deleteNode(selectedNodeId);
        return;
      }

      if (selectedEdgeId) {
        deleteEdge(selectedEdgeId);
      }
    },

    deleteNode(nodeId) {
      set((state) => {
        const nextEdges = state.edges.filter(
          (edge) => edge.source !== nodeId && edge.target !== nodeId,
        );

        return {
          nodes: state.nodes.filter((node) => node.id !== nodeId),
          edges: nextEdges,
          selectedNodeId:
            state.selectedNodeId === nodeId ? null : state.selectedNodeId,
          selectedEdgeId:
            state.selectedEdgeId &&
            nextEdges.some((edge) => edge.id === state.selectedEdgeId)
              ? state.selectedEdgeId
              : null,
        };
      });
    },

    deleteEdge(edgeId) {
      set((state) => ({
        edges: state.edges.filter((edge) => edge.id !== edgeId),
        selectedEdgeId:
          state.selectedEdgeId === edgeId ? null : state.selectedEdgeId,
      }));
    },

    updateNodeData(nodeId, patch) {
      set((state) => ({
        nodes: state.nodes.map((node) =>
          node.id === nodeId
            ? {
                ...node,
                data: {
                  ...node.data,
                  ...patch,
                },
              }
            : node,
        ),
      }));
    },

    updateNodeConfig(nodeId, patch) {
      set((state) => ({
        nodes: state.nodes.map((node) =>
          node.id === nodeId
            ? {
                ...node,
                data: {
                  ...node.data,
                  config: {
                    ...(node.data.config ?? {}),
                    ...patch,
                  },
                },
              }
            : node,
        ),
      }));
    },

    loadFromSchema(schema) {
      const nodes = schema.nodes.map((node) => ({
        id: node.id,
        type: getWorkflowNodeRenderType(node.type),
        position: node.position,
        data: {
          nodeType: node.type,
          label: node.label,
          description: node.description,
          config: {
            ...createDefaultWorkflowNodeConfig(node.type),
            ...(node.config ?? {}),
          },
        },
      })) satisfies WorkflowDesignerNode[];

      const edges = schema.edges.map((edge) => ({
        id: edge.id,
        source: edge.source,
        sourceHandle: edge.sourceHandle ?? null,
        target: edge.target,
        targetHandle: edge.targetHandle ?? null,
        type: edge.type ?? "smoothstep",
        markerEnd: {
          type: MarkerType.ArrowClosed,
        },
      })) satisfies WorkflowDesignerEdge[];

      set({
        nodes,
        edges,
        selectedNodeId: null,
        selectedEdgeId: null,
      });
    },

    resetDesigner() {
      set({
        nodes: createInitialNodes(),
        edges: createInitialEdges(),
        selectedNodeId: null,
        selectedEdgeId: null,
      });
    },
  }),
);

function createInitialNodes(): WorkflowDesignerNode[] {
  return [
    createWorkflowNode(getWorkflowNodeTemplate("Start"), "start_1", {
      x: 80,
      y: 220,
    }),
    createWorkflowNode(getWorkflowNodeTemplate("Prompt"), "prompt_1", {
      x: 320,
      y: 220,
    }),
    createWorkflowNode(getWorkflowNodeTemplate("LLM"), "llm_1", {
      x: 580,
      y: 220,
    }),
    createWorkflowNode(getWorkflowNodeTemplate("End"), "end_1", {
      x: 840,
      y: 220,
    }),
  ];
}

function createInitialEdges(): WorkflowDesignerEdge[] {
  return [
    createWorkflowEdge({
      id: "edge_start_prompt",
      source: "start_1",
      target: "prompt_1",
    }),
    createWorkflowEdge({
      id: "edge_prompt_llm",
      source: "prompt_1",
      target: "llm_1",
    }),
    createWorkflowEdge({
      id: "edge_llm_end",
      source: "llm_1",
      target: "end_1",
    }),
  ];
}

function createWorkflowNode(
  template: WorkflowNodeTemplate,
  id: string,
  position: XYPosition,
): WorkflowDesignerNode {
  return {
    id,
    type: getWorkflowNodeRenderType(template.type),
    position,
    data: {
      nodeType: template.type,
      label: template.label,
      description: template.description,
      config: createDefaultWorkflowNodeConfig(template.type),
    },
  };
}

function createWorkflowEdge(input: {
  id?: string;
  source: string;
  sourceHandle?: string | null;
  target: string;
  targetHandle?: string | null;
  selected?: boolean;
}): WorkflowDesignerEdge {
  return {
    id:
      input.id ??
      createEdgeId(
        input.source,
        input.target,
        input.sourceHandle,
        input.targetHandle,
      ),
    source: input.source,
    sourceHandle: input.sourceHandle,
    target: input.target,
    targetHandle: input.targetHandle,
    type: "smoothstep",
    markerEnd: {
      type: MarkerType.ArrowClosed,
    },
    selected: input.selected,
  };
}

function createNodeId(nodeType: WorkflowNodeType) {
  return `${nodeType.toLowerCase()}_${Date.now().toString(36)}_${Math.random()
    .toString(36)
    .slice(2, 8)}`;
}

function createEdgeId(
  source: string,
  target: string,
  sourceHandle?: string | null,
  targetHandle?: string | null,
) {
  return `edge_${source}_${sourceHandle ?? "source"}_${target}_${
    targetHandle ?? "target"
  }`;
}
