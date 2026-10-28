package workflowruntime

import (
	"context"
)

type EndExecutor struct{}

func NewEndExecutor() *EndExecutor {
	return &EndExecutor{}
}

func (e *EndExecutor) Type() WorkflowNodeType {
	return WorkflowNodeTypeEnd
}

func (e *EndExecutor) Validate(config WorkflowNodeConfig) error {
	return nil
}

func (e *EndExecutor) Execute(
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

	finalOutput := buildFinalOutput(input)
	inboundOutputs := buildInboundOutputList(input.InboundResults)

	output := JSONMap{
		"output":          finalOutput,
		"inbound_outputs": inboundOutputs,
		"variables":       input.ExecutionContext.SnapshotVariables(),
	}

	result := NewNodeExecutionResult(input.Node, output)
	result.Terminal = true
	result.Metadata = JSONMap{
		"executor":      "EndExecutor",
		"inbound_count": len(input.InboundResults),
	}

	return &result, nil
}

func buildFinalOutput(input NodeExecutionInput) JSONMap {
	if len(input.InboundResults) == 0 {
		return CloneJSONMap(input.Input)
	}

	if len(input.InboundResults) == 1 {
		return CloneJSONMap(input.InboundResults[0].Output)
	}

	output := JSONMap{}

	for _, inbound := range input.InboundResults {
		if inbound.NodeID == "" {
			continue
		}

		output[inbound.NodeID] = CloneJSONMap(inbound.Output)
	}

	return output
}

func buildInboundOutputList(inboundResults []NodeExecutionResult) []JSONMap {
	items := make([]JSONMap, 0, len(inboundResults))

	for _, inbound := range inboundResults {
		items = append(items, JSONMap{
			"node_id":   inbound.NodeID,
			"node_type": inbound.NodeType.String(),
			"output":    CloneJSONMap(inbound.Output),
		})
	}

	return items
}
