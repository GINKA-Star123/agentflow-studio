package workflowruntime

import "errors"

type ErrorCode string

const (
	ErrorCodeInvalidInput              ErrorCode = "WORKFLOWRUNTIME_INVALID_INPUT"
	ErrorCodeInvalidSchema             ErrorCode = "WORKFLOWRUNTIME_INVALID_SCHEMA"
	ErrorCodeInvalidExecutionContext   ErrorCode = "WORKFLOWRUNTIME_INVALID_EXECUTION_CONTEXT"
	ErrorCodeUnsupportedNodeType       ErrorCode = "WORKFLOWRUNTIME_UNSUPPORTED_NODE_TYPE"
	ErrorCodeWorkflowNotFound          ErrorCode = "WORKFLOWRUNTIME_WORKFLOW_NOT_FOUND"
	ErrorCodeRunNotFound               ErrorCode = "WORKFLOWRUNTIME_RUN_NOT_FOUND"
	ErrorCodeExecutorNotFound          ErrorCode = "WORKFLOWRUNTIME_EXECUTOR_NOT_FOUND"
	ErrorCodePermissionDenied          ErrorCode = "WORKFLOWRUNTIME_PERMISSION_DENIED"
	ErrorCodeInvalidDAG                ErrorCode = "WORKFLOWRUNTIME_INVALID_DAG"
	ErrorCodeCreateRunFailed           ErrorCode = "WORKFLOWRUNTIME_CREATE_RUN_FAILED"
	ErrorCodeUpdateRunFailed           ErrorCode = "WORKFLOWRUNTIME_UPDATE_RUN_FAILED"
	ErrorCodePromptRenderFailed        ErrorCode = "WORKFLOWRUNTIME_PROMPT_RENDER_FAILED"
	ErrorCodeNodeExecutionFailed       ErrorCode = "WORKFLOWRUNTIME_NODE_EXECUTION_FAILED"
	ErrorCodeRunAlreadyTerminal        ErrorCode = "RUNTIME_RUN_ALREADY_TERMINAL"
	ErrorCodeExecutorAlreadyRegistered ErrorCode = "WORKFLOWRUNTIME_EXECUTOR_ALREADY_REGISTERED"
	ErrorCodeCancelFailed              ErrorCode = "WORKFLOWRUNTIME_CANCEL_FAILED"
	ErrorCodeAIRuntimeError            ErrorCode = "WORKFLOWRUNTIME_AIRUNTIME_ERROR"
	ErrorCodeInvalidLLMConfig          ErrorCode = "WORKFLOWRUNTIME_INVALID_LLM_CONFIG"
)

var (
	ErrInvalidInput              = errors.New("runtime invalid input")
	ErrInvalidSchema             = errors.New("runtime invalid schema")
	ErrInvalidExecutionContext   = errors.New("runtime invalid execution context")
	ErrUnsupportedNodeType       = errors.New("runtime unsupported node type")
	ErrWorkflowNotFound          = errors.New("runtime workflow not found")
	ErrRunNotFound               = errors.New("runtime run not found")
	ErrExecutorNotFound          = errors.New("runtime executor not found")
	ErrPermissionDenied          = errors.New("runtime permission denied")
	ErrInvalidDAG                = errors.New("runtime invalid dag")
	ErrCreateRunFailed           = errors.New("runtime create run failed")
	ErrUpdateRunFailed           = errors.New("runtime update run failed")
	ErrPromptRenderFailed        = errors.New("runtime prompt render failed")
	ErrNodeExecutionFailed       = errors.New("runtime node execution failed")
	ErrRunAlreadyTerminal        = errors.New("runtime run already terminal")
	ErrExecutorAlreadyRegistered = errors.New("runtime executor already registered")
	ErrCancelFailed              = errors.New("runtime cancel failed")
	ErrAIRuntimeError            = errors.New("runtime ai runtime error")
	ErrInvalidLLMConfig          = errors.New("runtime invalid llm config")
)

type RuntimeError struct {
	Code    ErrorCode
	Message string
	Err     error
	Details any
}

func (e *RuntimeError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	if e.Err != nil {
		return e.Err.Error()
	}

	return string(e.Code)
}

func (e *RuntimeError) Unwrap() error {
	return e.Err
}

func NewRuntimeError(code ErrorCode, message string, err error) *RuntimeError {
	return &RuntimeError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func NewRuntimeErrorWithDetails(code ErrorCode, message string, err error, details any) *RuntimeError {
	return &RuntimeError{
		Code:    code,
		Message: message,
		Err:     err,
		Details: details,
	}
}
