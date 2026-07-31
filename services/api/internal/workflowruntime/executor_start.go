package workflowruntime

import (
	"context"
)

type StartExecutor struct{}

func NewStartExecutor() *StartExecutor {
	return &StartExecutor{}
}

func (e *StartExecutor) Type() WorkflowNodeType {
	return WorkflowNodeTypeStart
}

func (e *StartExecutor) Validate(config WorkflowNodeConfig) error {
	return nil
}

func (e *StartExecutor) Execute(
	ctx context.Context,
	input NodeExecutionInput,
) (*NodeExecutionResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := input.Validate(); err != nil {
		return nil, err
	}

	runInput := input.ExecutionContext.SnapshotInput()
	variables := input.ExecutionContext.SnapshotVariables()

	output := JSONMap{
		"input":     runInput,
		"variables": variables,
	}

	result := NewNodeExecutionResult(input.Node, output)
	result.Metadata = JSONMap{
		"executor": "StartExecutor",
	}

	return &result, nil
}
