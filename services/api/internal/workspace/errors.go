package workspace

import "errors"

type ErrorCode string

const (
	ErrorCodeInvalidInput             ErrorCode = "WORKSPACE_INVALID_INPUT"
	ErrorCodeInvalidRole              ErrorCode = "WORKSPACE_INVALID_ROLE"
	ErrorCodeNotFound                 ErrorCode = "WORKSPACE_NOT_FOUND"
	ErrorCodeForbidden                ErrorCode = "WORKSPACE_FORBIDDEN"
	ErrorCodeMemberNotFound           ErrorCode = "WORKSPACE_MEMBER_NOT_FOUND"
	ErrorCodeMemberAlreadyExists      ErrorCode = "WORKSPACE_MEMBER_ALREADY_EXISTS"
	ErrorCodeUserNotFound             ErrorCode = "WORKSPACE_USER_NOT_FOUND"
	ErrorCodeCreateFailed             ErrorCode = "WORKSPACE_CREATE_FAILED"
	ErrorCodePermissionDenied         ErrorCode = "WORKSPACE_PERMISSION_DENIED"
	ErrorCodeOwnerOperationNotAllowed ErrorCode = "WORKSPACE_OWNER_OPERATION_NOT_ALLOWED"
)

var (
	ErrInvalidInput             = errors.New("workspace invalid input")
	ErrInvalidRole              = errors.New("workspace invalid role")
	ErrNotFound                 = errors.New("workspace not found")
	ErrForbidden                = errors.New("workspace forbidden")
	ErrMemberNotFound           = errors.New("workspace member not found")
	ErrMemberAlreadyExists      = errors.New("workspace member already exists")
	ErrUserNotFound             = errors.New("workspace user not found")
	ErrCreateFailed             = errors.New("workspace create failed")
	ErrPermissionDenied         = errors.New("workspace permission denied")
	ErrOwnerOperationNotAllowed = errors.New("workspace owner operation not allowed")
)

type WorkspaceError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *WorkspaceError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	if e.Err != nil {
		return e.Err.Error()
	}

	return string(e.Code)
}

func (e *WorkspaceError) Unwrap() error {
	return e.Err
}

func NewWorkspaceError(code ErrorCode, message string, err error) *WorkspaceError {
	return &WorkspaceError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
