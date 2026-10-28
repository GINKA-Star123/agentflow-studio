package workflowruntime

import "strings"

const WorkflowSchemaVersion = "1.0"

type WorkflowNodeType string

const (
	WorkflowNodeTypeStart     WorkflowNodeType = "Start"
	WorkflowNodeTypeEnd       WorkflowNodeType = "End"
	WorkflowNodeTypePrompt    WorkflowNodeType = "Prompt"
	WorkflowNodeTypeLLM       WorkflowNodeType = "LLM"
	WorkflowNodeTypeCondition WorkflowNodeType = "Condition"
	WorkflowNodeTypeLoop      WorkflowNodeType = "Loop"
	WorkflowNodeTypeHTTP      WorkflowNodeType = "HTTP"
	WorkflowNodeTypeTool      WorkflowNodeType = "Tool"
	WorkflowNodeTypeMemory    WorkflowNodeType = "Memory"
	WorkflowNodeTypeRAG       WorkflowNodeType = "RAG"
)

var supportedWorkflowNodeTypes = map[WorkflowNodeType]struct{}{
	WorkflowNodeTypeStart:     {},
	WorkflowNodeTypeEnd:       {},
	WorkflowNodeTypePrompt:    {},
	WorkflowNodeTypeLLM:       {},
	WorkflowNodeTypeCondition: {},
	WorkflowNodeTypeLoop:      {},
	WorkflowNodeTypeHTTP:      {},
	WorkflowNodeTypeTool:      {},
	WorkflowNodeTypeMemory:    {},
	WorkflowNodeTypeRAG:       {},
}

var phase4ExecutableNodeTypes = map[WorkflowNodeType]struct{}{
	WorkflowNodeTypeStart:  {},
	WorkflowNodeTypeEnd:    {},
	WorkflowNodeTypePrompt: {},
	WorkflowNodeTypeLLM:    {},
}

type WorkflowSchema struct {
	SchemaVersion string                `json:"schema_version"`
	Name          string                `json:"name"`
	Summary       WorkflowSchemaSummary `json:"summary"`
	Nodes         []WorkflowSchemaNode  `json:"nodes"`
	Edges         []WorkflowSchemaEdge  `json:"edges"`
}

type WorkflowSchemaSummary struct {
	NodeCount  int `json:"node_count"`
	EdgeCount  int `json:"edge_count"`
	StartCount int `json:"start_count"`
	EndCount   int `json:"end_count"`
}

type WorkflowSchemaNode struct {
	ID          string               `json:"id"`
	Type        WorkflowNodeType     `json:"type"`
	Label       string               `json:"label"`
	Description string               `json:"description"`
	Position    WorkflowNodePosition `json:"position"`
	Config      WorkflowNodeConfig   `json:"config"`
}

type WorkflowNodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type WorkflowNodeConfig map[string]any

type WorkflowSchemaEdge struct {
	ID           string  `json:"id"`
	Source       string  `json:"source"`
	Target       string  `json:"target"`
	SourceHandle *string `json:"sourceHandle,omitempty"`
	TargetHandle *string `json:"targetHandle,omitempty"`
	Type         *string `json:"type,omitempty"`
}

func (t WorkflowNodeType) String() string {
	return string(t)
}

func (t WorkflowNodeType) IsSupported() bool {
	_, ok := supportedWorkflowNodeTypes[t]
	return ok
}

func (t WorkflowNodeType) IsExecutableInPhase4() bool {
	_, ok := phase4ExecutableNodeTypes[t]
	return ok
}

func IsWorkflowNodeType(value string) bool {
	nodeType := WorkflowNodeType(strings.TrimSpace(value))
	return nodeType.IsSupported()
}

func ParseWorkflowNodeType(value string) (WorkflowNodeType, error) {
	nodeType := WorkflowNodeType(strings.TrimSpace(value))
	if nodeType.IsSupported() {
		return nodeType, nil
	}

	return "", NewRuntimeErrorWithDetails(
		ErrorCodeUnsupportedNodeType,
		"不支持的 Workflow 节点类型",
		ErrUnsupportedNodeType,
		map[string]any{
			"value": value,
		},
	)
}
