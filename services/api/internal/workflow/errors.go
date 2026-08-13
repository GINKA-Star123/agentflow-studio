package workflow

import "errors"

type ErrorCode string

const (
	ErrorCodeInvalidInput     ErrorCode = "WORKFLOW_INVALID_INPUT"
	ErrorCodeNotFound         ErrorCode = "WORKFLOW_NOT_FOUND"
	ErrorCodePermissionDenied ErrorCode = "WORKFLOW_PERMISSION_DENIED"
	ErrorCodeCreateFailed     ErrorCode = "WORKFLOW_CREATE_FAILED"
	ErrorCodeUpdateFailed     ErrorCode = "WORKFLOW_UPDATE_FAILED"
)

var (
	ErrInvalidInput     = errors.New("workflow invalid input")
	ErrNotFound         = errors.New("workflow not found")
	ErrPermissionDenied = errors.New("workflow permission denied")
	ErrCreateFailed     = errors.New("workflow create failed")
	ErrUpdateFailed     = errors.New("workflow update failed")
)

type Error struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}

	if e.Err != nil {
		return e.Err.Error()
	}

	return string(e.Code)
}

func (e *Error) Unwrap() error {
	return e.Err
}

func NewError(code ErrorCode, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}
