package server

import (
	"agentflow-studio/services/api/internal/auth"
	"agentflow-studio/services/api/internal/config"
	"agentflow-studio/services/api/internal/handler"
	"agentflow-studio/services/api/internal/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewRouter(
	cfg *config.Config,
	logger *zap.Logger,
	authHandler *handler.AuthHandler,
	workspaceHandler *handler.WorkspaceHandler,
	jwtManager *auth.JWTManager,
) *gin.Engine {
	if cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.AccessLog(logger))

	healthHandler := handler.NewHealthHandler()

	router.GET("/healthz", healthHandler.Healthz)
	router.GET("/readyz", healthHandler.Readyz)

	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/health", healthHandler.Healthz)

		authGroup := apiV1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		protected := apiV1.Group("")
		protected.Use(middleware.AuthRequired(middleware.AuthRequiredConfig{
			JWTManager: jwtManager,
		}))
		{
			protected.GET("/auth/me", authHandler.Me)

			workspaceGroup := protected.Group("/workspaces")
			{
				workspaceGroup.GET("", workspaceHandler.List)
				workspaceGroup.POST("", workspaceHandler.Create)

				workspaceGroup.GET("/:workspace_id/members", workspaceHandler.ListMembers)
				workspaceGroup.POST("/:workspace_id/members", workspaceHandler.AddMember)
				workspaceGroup.PATCH("/:workspace_id/members/:user_id/role", workspaceHandler.UpdateMemberRole)
				workspaceGroup.DELETE("/:workspace_id/members/:user_id", workspaceHandler.RemoveMember)
			}
		}
	}

	return router
}
