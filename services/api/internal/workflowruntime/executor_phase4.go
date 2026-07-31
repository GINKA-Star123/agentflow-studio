package workflowruntime

import "agentflow-studio/services/api/internal/airuntime"

func RegisterPhase4Executors(
	registry *ExecutorRegistry,
	aiRuntimeClient *airuntime.Client,
) error {
	if registry == nil {
		return NewRuntimeError(
			ErrorCodeInvalidInput,
			"执行器注册表不能为空",
			ErrInvalidInput,
		)
	}

	if err := registry.Register(NewStartExecutor()); err != nil {
		return err
	}

	if err := registry.Register(NewEndExecutor()); err != nil {
		return err
	}

	if err := registry.Register(NewPromptExecutor()); err != nil {
		return err
	}

	if err := registry.Register(NewLLMExecutor(aiRuntimeClient)); err != nil {
		return err
	}

	return nil
}

func NewPhase4ExecutorRegistry(aiRuntimeClient *airuntime.Client) *ExecutorRegistry {
	registry := NewExecutorRegistry()

	registry.MustRegister(NewStartExecutor())
	registry.MustRegister(NewEndExecutor())
	registry.MustRegister(NewPromptExecutor())
	registry.MustRegister(NewLLMExecutor(aiRuntimeClient))

	return registry
}
