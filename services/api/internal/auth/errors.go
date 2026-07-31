package auth

import "errors"

type ErrorCode string

const (
	ErrorCodeInvalidInput       ErrorCode = "INVALID_INPUT"
	ErrorCodeEmailAlreadyExists ErrorCode = "EMAIL_ALREADY_EXISTS"
	ErrorCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrorCodeUserDisabled       ErrorCode = "USER_DISABLED"
	ErrorCodeWeakPassword       ErrorCode = "WEAK_PASSWORD"
	ErrorCodePasswordTooLong    ErrorCode = "PASSWORD_TOO_LONG"
	ErrorCodeMissingToken       ErrorCode = "MISSING_TOKEN"
	ErrorCodeInvalidToken       ErrorCode = "INVALID_TOKEN"
	ErrorCodeExpiredToken       ErrorCode = "EXPIRED_TOKEN"
	ErrorCodeTokenSigningFailed ErrorCode = "TOKEN_SIGNING_FAILED"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user is disabled")
	ErrWeakPassword       = errors.New("password is too weak")
	ErrPasswordTooLong    = errors.New("password is too long")
	ErrMissingToken       = errors.New("missing token")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("expired token")
	ErrTokenSigningFailed = errors.New("token signing failed")
)

type AuthError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *AuthError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

func (e *AuthError) Unwrap() error {
	return e.Err
}

func NewAuthError(code ErrorCode, message string, err error) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
