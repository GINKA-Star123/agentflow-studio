import type { Edge, Node } from "reactflow";

export const workflowNodeTypes = [
  "Start",
  "End",
  "Prompt",
  "LLM",
  "Condition",
  "Loop",
  "HTTP",
  "Tool",
  "Memory",
  "RAG",
] as const;

export type WorkflowNodeType = (typeof workflowNodeTypes)[number];

export type WorkflowNodeRenderType = Lowercase<WorkflowNodeType>;

export const workflowNodeGroups = ["基础", "AI", "控制", "集成", "数据"] as const;

export type WorkflowNodeGroup = (typeof workflowNodeGroups)[number];

export type WorkflowNodeTemplate = {
  type: WorkflowNodeType;
  label: string;
  description: string;
  group: WorkflowNodeGroup;
};

export type WorkflowNodeConfig = {
  promptTemplate?: string;
  variables?: string;
  provider?: string;
  model?: string;
  temperature?: number;
  maxTokens?: number;
  systemPrompt?: string;
};

export type WorkflowNodeData = {
  nodeType: WorkflowNodeType;
  label: string;
  description: string;
  config: WorkflowNodeConfig;
};

export type WorkflowDesignerNode = Node<WorkflowNodeData>;

export type WorkflowDesignerEdge = Edge;

export const WORKFLOW_NODE_DRAG_TYPE = "application/agentflow-workflow-node";

export const workflowNodeTemplates: WorkflowNodeTemplate[] = [
  {
    type: "Start",
    label: "Start",
    description: "流程入口",
    group: "基础",
  },
  {
    type: "End",
    label: "End",
    description: "流程出口",
    group: "基础",
  },
  {
    type: "Prompt",
    label: "Prompt",
    description: "模板渲染",
    group: "AI",
  },
  {
    type: "LLM",
    label: "LLM",
    description: "模型调用",
    group: "AI",
  },
  {
    type: "Condition",
    label: "Condition",
    description: "分支判断",
    group: "控制",
  },
  {
    type: "Loop",
    label: "Loop",
    description: "循环控制",
    group: "控制",
  },
  {
    type: "HTTP",
    label: "HTTP",
    description: "外部请求",
    group: "集成",
  },
  {
    type: "Tool",
    label: "Tool",
    description: "工具调用",
    group: "集成",
  },
  {
    type: "Memory",
    label: "Memory",
    description: "记忆读写",
    group: "数据",
  },
  {
    type: "RAG",
    label: "RAG",
    description:"知识检索",
    group: "数据",
  },
];

export const workflowNodeTemplateMap = workflowNodeTemplates.reduce<
  Record<WorkflowNodeType, WorkflowNodeTemplate>
>((map, template) => {
  map[template.type] = template;
  return map;
}, {} as Record<WorkflowNodeType, WorkflowNodeTemplate>);

export function getWorkflowNodeTemplate(type: WorkflowNodeType) {
  return workflowNodeTemplateMap[type];
}

export function getWorkflowNodeRenderType(
  type: WorkflowNodeType,
): WorkflowNodeRenderType {
  return type.toLowerCase() as WorkflowNodeRenderType;
}

export function isWorkflowNodeType(value: string): value is WorkflowNodeType {
  return workflowNodeTypes.includes(value as WorkflowNodeType);
}

export function createDefaultWorkflowNodeConfig(
  nodeType: WorkflowNodeType,
): WorkflowNodeConfig {
  switch (nodeType) {
    case "Prompt":
      return {
        promptTemplate: "",
        variables: "",
      };

    case "LLM":
      return {
        provider: "",
        model: "",
        temperature: 0.7,
        maxTokens: 1024,
        systemPrompt: "",
      };

    default:
      return {};
  }
}