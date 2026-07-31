package workflowruntime

func RegisterBasicExecutors(registry *ExecutorRegistry) error {
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

	return nil
}

func NewBasicExecutorRegistry() *ExecutorRegistry {
	registry := NewExecutorRegistry()

	registry.MustRegister(NewStartExecutor())
	registry.MustRegister(NewEndExecutor())
	registry.MustRegister(NewPromptExecutor())

	return registry
}
