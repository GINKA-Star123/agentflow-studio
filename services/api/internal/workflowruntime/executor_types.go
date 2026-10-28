package workflowruntime

import (
	"context"
	"time"
)

type NodeExecutor interface {
	Type() WorkflowNodeType
	Validate(config WorkflowNodeConfig) error
	Execute(ctx context.Context, input NodeExecutionInput) (*NodeExecutionResult, error)
}

type NodeExecutionInput struct {
	ExecutionContext *ExecutionContext
	Node             WorkflowSchemaNode
	Input            JSONMap
	InboundResults   []NodeExecutionResult
	StartedAt        time.Time
}

func NewNodeExecutionInput(
	executionContext *ExecutionContext,
	node WorkflowSchemaNode,
	input JSONMap,
	inboundResults []NodeExecutionResult,
) NodeExecutionInput {
	return NodeExecutionInput{
		ExecutionContext: executionContext,
		Node:             node,
		Input:            CloneJSONMap(input),
		InboundResults:   cloneNodeExecutionResults(inboundResults),
		StartedAt:        time.Now().UTC(),
	}
}

func (i NodeExecutionInput) Validate() error {
	if err := i.ExecutionContext.Validate(); err != nil {
		return err
	}

	if i.Node.ID == "" {
		return NewRuntimeErrorWithDetails(
			ErrorCodeInvalidInput,
			"节点执行输入缺少 node_id",
			ErrInvalidInput,
			map[string]any{"field": "node.id"},
		)
	}

	if !i.Node.Type.IsSupported() {
		return NewRuntimeErrorWithDetails(
			ErrorCodeUnsupportedNodeType,
			"不支持的节点类型",
			ErrUnsupportedNodeType,
			map[string]any{
				"node_id":   i.Node.ID,
				"node_type": i.Node.Type.String(),
			},
		)
	}

	return nil
}

func (i NodeExecutionInput) Config() WorkflowNodeConfig {
	if i.Node.Config == nil {
		return WorkflowNodeConfig{}
	}

	return i.Node.Config
}

func (i NodeExecutionInput) CloneInboundResults() []NodeExecutionResult {
	return cloneNodeExecutionResults(i.InboundResults)
}

type NodeExecutionResult struct {
	NodeID     string           `json:"node_id"`
	NodeType   WorkflowNodeType `json:"node_type"`
	Output     JSONMap          `json:"output"`
	Variables  JSONMap          `json:"variables,omitempty"`
	TokenUsage *TokenUsage      `json:"token_usage,omitempty"`
	Metadata   JSONMap          `json:"metadata,omitempty"`
	Terminal   bool             `json:"terminal"`
}

func NewNodeExecutionResult(node WorkflowSchemaNode, output JSONMap) NodeExecutionResult {
	return NodeExecutionResult{
		NodeID:    node.ID,
		NodeType:  node.Type,
		Output:    CloneJSONMap(output),
		Variables: JSONMap{},
		Metadata:  JSONMap{},
		Terminal:  node.Type == WorkflowNodeTypeEnd,
	}
}

func (r NodeExecutionResult) Clone() NodeExecutionResult {
	copied := NodeExecutionResult{
		NodeID:    r.NodeID,
		NodeType:  r.NodeType,
		Output:    CloneJSONMap(r.Output),
		Variables: CloneJSONMap(r.Variables),
		Metadata:  CloneJSONMap(r.Metadata),
		Terminal:  r.Terminal,
	}

	if r.TokenUsage != nil {
		tokenUsage := r.TokenUsage.Normalize()
		copied.TokenUsage = &tokenUsage
	}

	return copied
}

func (r NodeExecutionResult) OutputValue(key string) (any, bool) {
	if r.Output == nil {
		return nil, false
	}

	value, ok := r.Output[key]
	return value, ok
}

func (r NodeExecutionResult) HasOutputKey(key string) bool {
	_, ok := r.OutputValue(key)
	return ok
}

func (r NodeExecutionResult) ApplyToContext(executionContext *ExecutionContext) error {
	if executionContext == nil {
		return NewRuntimeError(
			ErrorCodeInvalidExecutionContext,
			"执行上下文不能为空",
			ErrInvalidExecutionContext,
		)
	}

	if err := executionContext.SetNodeOutput(r.NodeID, r); err != nil {
		return err
	}

	executionContext.MergeVariables(r.Variables)
	return nil
}

type TokenUsage struct {
	Provider     string `json:"provider,omitempty"`
	Model        string `json:"model,omitempty"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	TotalTokens  int    `json:"total_tokens"`
}

func (u TokenUsage) Normalize() TokenUsage {
	if u.TotalTokens <= 0 {
		u.TotalTokens = u.InputTokens + u.OutputTokens
	}

	return u
}

func cloneNodeExecutionResults(values []NodeExecutionResult) []NodeExecutionResult {
	result := make([]NodeExecutionResult, 0, len(values))

	for _, value := range values {
		result = append(result, value.Clone())
	}

	return result
}
