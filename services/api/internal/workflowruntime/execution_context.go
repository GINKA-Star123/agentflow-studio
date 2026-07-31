package workflowruntime

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

type JSONMap map[string]any

type ExecutionContextInput struct {
	WorkspaceID uuid.UUID
	WorkflowID  uuid.UUID
	RunID       uuid.UUID
	UserID      uuid.UUID
	Input       JSONMap
	Variables   JSONMap
	TraceID     string
}

type ExecutionContext struct {
	WorkspaceID uuid.UUID                      `json:"workspace_id"`
	WorkflowID  uuid.UUID                      `json:"workflow_id"`
	RunID       uuid.UUID                      `json:"run_id"`
	UserID      uuid.UUID                      `json:"user_id"`
	Input       JSONMap                        `json:"input"`
	Variables   JSONMap                        `json:"variables"`
	NodeOutputs map[string]NodeExecutionResult `json:"node_outputs"`
	TraceID     string                         `json:"trace_id"`
}

func NewExecutionContext(input ExecutionContextInput) *ExecutionContext {
	return &ExecutionContext{
		WorkspaceID: input.WorkspaceID,
		WorkflowID:  input.WorkflowID,
		RunID:       input.RunID,
		UserID:      input.UserID,
		Input:       CloneJSONMap(input.Input),
		Variables:   CloneJSONMap(input.Variables),
		NodeOutputs: map[string]NodeExecutionResult{},
		TraceID:     strings.TrimSpace(input.TraceID),
	}
}

func (c *ExecutionContext) Validate() error {
	if c == nil {
		return NewRuntimeError(
			ErrorCodeInvalidExecutionContext,
			"执行上下文不能为空",
			ErrInvalidExecutionContext,
		)
	}

	if c.WorkspaceID == uuid.Nil {
		return NewRuntimeErrorWithDetails(
			ErrorCodeInvalidExecutionContext,
			"执行上下文缺少 workspace_id",
			ErrInvalidExecutionContext,
			map[string]any{"field": "workspace_id"},
		)
	}

	if c.WorkflowID == uuid.Nil {
		return NewRuntimeErrorWithDetails(
			ErrorCodeInvalidExecutionContext,
			"执行上下文缺少 workflow_id",
			ErrInvalidExecutionContext,
			map[string]any{"field": "workflow_id"},
		)
	}

	if c.RunID == uuid.Nil {
		return NewRuntimeErrorWithDetails(
			ErrorCodeInvalidExecutionContext,
			"执行上下文缺少 run_id",
			ErrInvalidExecutionContext,
			map[string]any{"field": "run_id"},
		)
	}

	if c.UserID == uuid.Nil {
		return NewRuntimeErrorWithDetails(
			ErrorCodeInvalidExecutionContext,
			"执行上下文缺少 user_id",
			ErrInvalidExecutionContext,
			map[string]any{"field": "user_id"},
		)
	}

	if c.Input == nil {
		c.Input = JSONMap{}
	}

	if c.Variables == nil {
		c.Variables = JSONMap{}
	}

	if c.NodeOutputs == nil {
		c.NodeOutputs = map[string]NodeExecutionResult{}
	}

	return nil
}

func (c *ExecutionContext) SetVariable(key string, value any) {
	if c == nil {
		return
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return
	}

	if c.Variables == nil {
		c.Variables = JSONMap{}
	}

	c.Variables[key] = value
}

func (c *ExecutionContext) GetVariable(key string) (any, bool) {
	if c == nil || c.Variables == nil {
		return nil, false
	}

	value, ok := c.Variables[strings.TrimSpace(key)]
	return value, ok
}

func (c *ExecutionContext) MergeVariables(variables JSONMap) {
	if c == nil {
		return
	}

	for key, value := range variables {
		c.SetVariable(key, value)
	}
}

func (c *ExecutionContext) SetNodeOutput(nodeID string, result NodeExecutionResult) error {
	if c == nil {
		return NewRuntimeError(
			ErrorCodeInvalidExecutionContext,
			"执行上下文不能为空",
			ErrInvalidExecutionContext,
		)
	}

	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return NewRuntimeErrorWithDetails(
			ErrorCodeInvalidExecutionContext,
			"节点输出缺少 node_id",
			ErrInvalidExecutionContext,
			map[string]any{"field": "node_id"},
		)
	}

	if c.NodeOutputs == nil {
		c.NodeOutputs = map[string]NodeExecutionResult{}
	}

	copied := result.Clone()
	if copied.NodeID == "" {
		copied.NodeID = nodeID
	}

	c.NodeOutputs[nodeID] = copied
	return nil
}

func (c *ExecutionContext) GetNodeOutput(nodeID string) (NodeExecutionResult, bool) {
	if c == nil || c.NodeOutputs == nil {
		return NodeExecutionResult{}, false
	}

	result, ok := c.NodeOutputs[strings.TrimSpace(nodeID)]
	return result.Clone(), ok
}

func (c *ExecutionContext) SnapshotInput() JSONMap {
	if c == nil {
		return JSONMap{}
	}

	return CloneJSONMap(c.Input)
}

func (c *ExecutionContext) SnapshotVariables() JSONMap {
	if c == nil {
		return JSONMap{}
	}

	return CloneJSONMap(c.Variables)
}

func (c *ExecutionContext) SnapshotNodeOutputs() map[string]JSONMap {
	result := map[string]JSONMap{}

	if c == nil {
		return result
	}

	for nodeID, nodeResult := range c.NodeOutputs {
		result[nodeID] = CloneJSONMap(nodeResult.Output)
	}

	return result
}

func CloneJSONMap(value JSONMap) JSONMap {
	if value == nil {
		return JSONMap{}
	}

	raw, err := json.Marshal(value)
	if err == nil {
		var copied map[string]any
		if err := json.Unmarshal(raw, &copied); err == nil {
			return JSONMap(copied)
		}
	}

	copied := JSONMap{}
	for key, item := range value {
		copied[key] = item
	}

	return copied
}
