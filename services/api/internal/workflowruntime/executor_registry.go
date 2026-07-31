package workflowruntime

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

type ExecutorRegistry struct {
	mu        sync.RWMutex
	executors map[WorkflowNodeType]NodeExecutor
}

func NewExecutorRegistry() *ExecutorRegistry {
	return &ExecutorRegistry{
		executors: map[WorkflowNodeType]NodeExecutor{},
	}
}

func (r *ExecutorRegistry) Register(executor NodeExecutor) error {
	if executor == nil {
		return NewRuntimeError(
			ErrorCodeInvalidInput,
			"节点执行器不能为空",
			ErrInvalidInput,
		)
	}

	nodeType := executor.Type()
	if !nodeType.IsSupported() {
		return NewRuntimeErrorWithDetails(
			ErrorCodeUnsupportedNodeType,
			"不支持的节点执行器类型",
			ErrUnsupportedNodeType,
			map[string]any{
				"node_type": nodeType.String(),
			},
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.executors[nodeType]; exists {
		return NewRuntimeErrorWithDetails(
			ErrorCodeExecutorAlreadyRegistered,
			"节点执行器已注册",
			ErrExecutorAlreadyRegistered,
			map[string]any{
				"node_type": nodeType.String(),
			},
		)
	}

	r.executors[nodeType] = executor
	return nil
}

func (r *ExecutorRegistry) MustRegister(executor NodeExecutor) {
	if err := r.Register(executor); err != nil {
		panic(err)
	}
}

func (r *ExecutorRegistry) Get(nodeType WorkflowNodeType) (NodeExecutor, error) {
	if r == nil {
		return nil, NewRuntimeError(
			ErrorCodeExecutorNotFound,
			"节点执行器注册表未初始化",
			ErrExecutorNotFound,
		)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	executor, exists := r.executors[nodeType]
	if !exists {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeExecutorNotFound,
			"节点执行器不存在",
			ErrExecutorNotFound,
			map[string]any{
				"node_type": nodeType.String(),
			},
		)
	}

	return executor, nil
}

func (r *ExecutorRegistry) Has(nodeType WorkflowNodeType) bool {
	if r == nil {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.executors[nodeType]
	return exists
}

func (r *ExecutorRegistry) ListTypes() []WorkflowNodeType {
	if r == nil {
		return []WorkflowNodeType{}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]WorkflowNodeType, 0, len(r.executors))
	for nodeType := range r.executors {
		result = append(result, nodeType)
	}

	sort.Slice(result, func(i int, j int) bool {
		return result[i].String() < result[j].String()
	})

	return result
}

func (r *ExecutorRegistry) ValidateNode(node WorkflowSchemaNode) error {
	executor, err := r.Get(node.Type)
	if err != nil {
		return err
	}

	if err := executor.Validate(node.Config); err != nil {
		return NewRuntimeErrorWithDetails(
			ErrorCodeInvalidInput,
			"节点配置校验失败",
			err,
			map[string]any{
				"node_id":   node.ID,
				"node_type": node.Type.String(),
			},
		)
	}

	return nil
}

func (r *ExecutorRegistry) ValidateSchemaExecutors(schema *WorkflowSchema) error {
	if schema == nil {
		return NewRuntimeError(
			ErrorCodeInvalidSchema,
			"Workflow Schema 不能为空",
			ErrInvalidSchema,
		)
	}

	for _, node := range schema.Nodes {
		if !node.Type.IsExecutableInPhase4() {
			return NewRuntimeErrorWithDetails(
				ErrorCodeUnsupportedNodeType,
				"当前阶段暂不支持执行该节点类型",
				ErrUnsupportedNodeType,
				map[string]any{
					"node_id":   node.ID,
					"node_type": node.Type.String(),
				},
			)
		}

		if err := r.ValidateNode(node); err != nil {
			return err
		}
	}

	return nil
}

func (r *ExecutorRegistry) ExecuteNode(
	ctx context.Context,
	input NodeExecutionInput,
) (*NodeExecutionResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	executor, err := r.Get(input.Node.Type)
	if err != nil {
		return nil, err
	}

	if err := executor.Validate(input.Config()); err != nil {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeInvalidInput,
			"节点配置校验失败",
			err,
			map[string]any{
				"node_id":   input.Node.ID,
				"node_type": input.Node.Type.String(),
			},
		)
	}

	result, err := executor.Execute(ctx, input)
	if err != nil {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeNodeExecutionFailed,
			"节点执行失败",
			err,
			map[string]any{
				"node_id":   input.Node.ID,
				"node_type": input.Node.Type.String(),
			},
		)
	}

	if result == nil {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeNodeExecutionFailed,
			"节点执行结果不能为空",
			ErrNodeExecutionFailed,
			map[string]any{
				"node_id":   input.Node.ID,
				"node_type": input.Node.Type.String(),
			},
		)
	}

	if result.NodeID == "" {
		result.NodeID = input.Node.ID
	}

	if result.NodeType == "" {
		result.NodeType = input.Node.Type
	}

	result.Output = CloneJSONMap(result.Output)
	result.Variables = CloneJSONMap(result.Variables)
	result.Metadata = CloneJSONMap(result.Metadata)

	if result.TokenUsage != nil {
		tokenUsage := result.TokenUsage.Normalize()
		result.TokenUsage = &tokenUsage
	}

	if result.NodeID != input.Node.ID {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeNodeExecutionFailed,
			"节点执行结果 node_id 与当前节点不一致",
			ErrNodeExecutionFailed,
			map[string]any{
				"expected": input.Node.ID,
				"actual":   result.NodeID,
			},
		)
	}

	if result.NodeType != input.Node.Type {
		return nil, NewRuntimeErrorWithDetails(
			ErrorCodeNodeExecutionFailed,
			"节点执行结果 node_type 与当前节点不一致",
			ErrNodeExecutionFailed,
			map[string]any{
				"expected": input.Node.Type.String(),
				"actual":   result.NodeType.String(),
			},
		)
	}

	return result, nil
}

func (r *ExecutorRegistry) String() string {
	return fmt.Sprintf("ExecutorRegistry%v", r.ListTypes())
}
