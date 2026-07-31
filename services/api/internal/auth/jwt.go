package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeAccess TokenType = "access"
)

type JWTConfig struct {
	Secret         string
	Issuer         string
	AccessTokenTTL time.Duration
}

type JWTManager struct {
	secret         []byte
	issuer         string
	accessTokenTTL time.Duration
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Type   TokenType `json:"type"`

	jwt.RegisteredClaims
}

func NewJWTManager(cfg JWTConfig) (*JWTManager, error) {
	if strings.TrimSpace(cfg.Secret) == "" {
		return nil, NewAuthError(
			ErrorCodeInvalidToken,
			"JWT_SECRET 不能为空",
			ErrInvalidToken,
		)
	}

	if strings.TrimSpace(cfg.Issuer) == "" {
		return nil, NewAuthError(
			ErrorCodeInvalidToken,
			"JWT_ISSUER 不能为空",
			ErrInvalidToken,
		)
	}

	if cfg.AccessTokenTTL <= 0 {
		return nil, NewAuthError(
			ErrorCodeInvalidToken,
			"JWT_ACCESS_TOKEN_TTL 必须大于 0",
			ErrInvalidToken,
		)
	}

	return &JWTManager{
		secret:         []byte(cfg.Secret),
		issuer:         cfg.Issuer,
		accessTokenTTL: cfg.AccessTokenTTL,
	}, nil
}

func (m *JWTManager) GenerateAccessToken(userID uuid.UUID, email string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.accessTokenTTL)

	claims := Claims{
		UserID: userID,
		Email:  email,
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, NewAuthError(
			ErrorCodeTokenSigningFailed,
			"Token 签发失败",
			err,
		)
	}

	return tokenString, expiresAt, nil
}

func (m *JWTManager) VerifyAccessToken(tokenString string) (*Claims, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, NewAuthError(
			ErrorCodeMissingToken,
			"缺少 Token",
			ErrMissingToken,
		)
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}

			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, NewAuthError(
				ErrorCodeExpiredToken,
				"Token 已过期",
				ErrExpiredToken,
			)
		}

		return nil, NewAuthError(
			ErrorCodeInvalidToken,
			"Token 无效",
			ErrInvalidToken,
		)
	}

	if token == nil || !token.Valid {
		return nil, NewAuthError(
			ErrorCodeInvalidToken,
			"Token 无效",
			ErrInvalidToken,
		)
	}

	if claims.Type != TokenTypeAccess {
		return nil, NewAuthError(
			ErrorCodeInvalidToken,
			"Token 类型无效",
			ErrInvalidToken,
		)
	}

	if claims.UserID == uuid.Nil {
		return nil, NewAuthError(
			ErrorCodeInvalidToken,
			"Token 用户 ID 无效",
			ErrInvalidToken,
		)
	}

	return claims, nil
}

func ExtractBearerToken(authorization string) (string, error) {
	authorization = strings.TrimSpace(authorization)

	if authorization == "" {
		return "", NewAuthError(
			ErrorCodeMissingToken,
			"缺少 Authorization 请求头",
			ErrMissingToken,
		)
	}

	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 {
		return "", NewAuthError(
			ErrorCodeInvalidToken,
			"Authorization 格式无效",
			ErrInvalidToken,
		)
	}

	scheme := strings.TrimSpace(parts[0])
	token := strings.TrimSpace(parts[1])

	if !strings.EqualFold(scheme, "Bearer") || token == "" {
		return "", NewAuthError(
			ErrorCodeInvalidToken,
			"Authorization 必须使用 Bearer Token",
			ErrInvalidToken,
		)
	}

	return token, nil
}
