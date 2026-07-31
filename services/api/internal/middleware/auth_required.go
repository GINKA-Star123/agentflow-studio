package middleware

import (
	"agentflow-studio/services/api/internal/auth"
	"agentflow-studio/services/api/internal/requestctx"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthRequiredConfig struct {
	JWTManager *auth.JWTManager
}

func AuthRequired(cfg AuthRequiredConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.JWTManager == nil {
			abortAuthError(
				c,
				http.StatusInternalServerError,
				"INTERNAL",
				"认证服务未初始化",
			)
			return
		}
		authorization := c.GetHeader("Authorization")

		tokenString, err := auth.ExtractBearerToken(authorization)
		if err != nil {
			abortAuthError(
				c,
				http.StatusUnauthorized,
				"MISSING_TOKEN",
				"缺少或无效的Authorization Bearer Token",
			)
			return
		}

		claims, err := cfg.JWTManager.VerifyAccessToken(tokenString)
		if err != nil {
			abortAuthError(
				c,
				http.StatusUnauthorized,
				"INVALID_TOKEN",
				"无效的Token",
			)
			return
		}

		requestctx.SetCurrentUser(c, requestctx.CurrentUser{
			ID:    claims.UserID,
			Email: claims.Email,
		})
		c.Next()
	}
}

func abortAuthError(c *gin.Context, status int, code string, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
		"request_id": GetRequestID(c),
	})
}
