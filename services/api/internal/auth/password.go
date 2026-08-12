package auth

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength      = 8
	MaxBcryptPasswordBytes = 72
)

func ValidatePassword(password string) error {
	password = strings.TrimSpace(password)

	if len(password) < MinPasswordLength {
		return NewAuthError(
			ErrorCodeWeakPassword,
			"密码长度不能少于8位",
			ErrWeakPassword,
		)
	}
	if len([]byte(password)) > MaxBcryptPasswordBytes {
		return NewAuthError(
			ErrorCodePasswordTooLong,
			"密码长度不能超过72字节",
			ErrPasswordTooLong,
		)
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func ComparePassword(passwordHash string, password string) error {
	if passwordHash == "" || password == "" {
		return NewAuthError(
			ErrorCodeInvalidCredentials,
			"邮箱或密码错误",
			ErrInvalidCredentials,
		)

	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return NewAuthError(
			ErrorCodeInvalidCredentials,
			"邮箱或密码错误",
			ErrInvalidCredentials,
		)
	}

	return nil
}
