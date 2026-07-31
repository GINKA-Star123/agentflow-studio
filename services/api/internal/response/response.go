package response

import (
	"agentflow-studio/services/api/internal/auth"
	"agentflow-studio/services/api/internal/middleware"
	"agentflow-studio/services/api/internal/workspace"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Body struct {
	Data      any       `json:"data,omitempty"`
	Error     *APIError `json:"error,omitempty"`
	RequestID string    `json:"request_id"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func OK(c *gin.Context, data any) {
	JSON(c, http.StatusOK, data)
}

func Created(c *gin.Context, data any) {
	JSON(c, http.StatusCreated, data)
}

func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, Body{
		Data:      data,
		RequestID: middleware.GetRequestID(c),
	})
}

func Fail(c *gin.Context, status int, code string, message string, details any) {
	c.JSON(status, Body{
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: middleware.GetRequestID(c),
	})
}

func BindError(c *gin.Context, err error) {
	Fail(
		c,
		http.StatusBadRequest,
		"INVALID_ARGUMENT",
		"请求参数无效",
		gin.H{
			"reason": err.Error(),
		},
	)
}

func FromError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var authErr *auth.AuthError
	if errors.As(err, &authErr) {
		fromAuthError(c, authErr)
		return
	}

	var workspaceErr *workspace.WorkspaceError
	if errors.As(err, &workspaceErr) {
		fromWorkspaceError(c, workspaceErr)
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		Fail(
			c,
			http.StatusNotFound,
			"NOT FOUNT",
			"资源不存在",
			nil,
		)
		return
	}

	Fail(
		c,
		http.StatusInternalServerError,
		"INTERNAL",
		"服务内部错误",
		nil,
	)
}

func fromAuthError(c *gin.Context, err *auth.AuthError) {
	switch err.Code {
	case auth.ErrorCodeInvalidInput,
		auth.ErrorCodeWeakPassword,
		auth.ErrorCodePasswordTooLong:
		Fail(
			c,
			http.StatusBadRequest,
			string(err.Code),
			err.Error(),
			nil,
		)

	case auth.ErrorCodeEmailAlreadyExists:
		Fail(
			c,
			http.StatusConflict,
			string(err.Code),
			err.Error(),
			nil,
		)

	case auth.ErrorCodeInvalidCredentials,
		auth.ErrorCodeMissingToken,
		auth.ErrorCodeInvalidToken,
		auth.ErrorCodeExpiredToken:
		Fail(
			c,
			http.StatusUnauthorized,
			string(err.Code),
			err.Error(),
			nil,
		)

	case auth.ErrorCodeUserDisabled:
		Fail(
			c,
			http.StatusForbidden,
			string(err.Code),
			err.Error(),
			nil,
		)

	case auth.ErrorCodeTokenSigningFailed:
		Fail(
			c,
			http.StatusInternalServerError,
			string(err.Code),
			err.Error(),
			nil,
		)

	default:
		Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL",
			"服务内部错误",
			nil,
		)
	}
}

func fromWorkspaceError(c *gin.Context, err *workspace.WorkspaceError) {
	switch err.Code {
	case workspace.ErrorCodeInvalidInput,
		workspace.ErrorCodeInvalidRole:
		Fail(
			c,
			http.StatusBadRequest,
			string(err.Code),
			err.Error(),
			nil,
		)

	case workspace.ErrorCodeNotFound,
		workspace.ErrorCodeMemberNotFound,
		workspace.ErrorCodeUserNotFound:
		Fail(
			c,
			http.StatusNotFound,
			string(err.Code),
			err.Error(),
			nil,
		)

	case workspace.ErrorCodeMemberAlreadyExists:
		Fail(
			c,
			http.StatusConflict,
			string(err.Code),
			err.Error(),
			nil,
		)
	case workspace.ErrorCodeForbidden,
		workspace.ErrorCodePermissionDenied,
		workspace.ErrorCodeOwnerOperationNotAllowed:
		Fail(
			c,
			http.StatusForbidden,
			string(err.Code),
			err.Error(),
			nil,
		)

	case workspace.ErrorCodeCreateFailed:
		Fail(
			c,
			http.StatusInternalServerError,
			string(err.Code),
			err.Error(),
			nil,
		)

	default:
		Fail(
			c,
			http.StatusInternalServerError,
			"INTERNAL",
			"服务内部错误",
			nil,
		)
	}
}
