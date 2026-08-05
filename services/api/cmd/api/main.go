package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agentflow-studio/services/api/internal/airuntime"
	"agentflow-studio/services/api/internal/auth"
	"agentflow-studio/services/api/internal/config"
	appdb "agentflow-studio/services/api/internal/database"
	"agentflow-studio/services/api/internal/handler"
	"agentflow-studio/services/api/internal/repository"
	"agentflow-studio/services/api/internal/server"
	"agentflow-studio/services/api/internal/service"
	"agentflow-studio/services/api/internal/workflowruntime"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger, err := newLogger(cfg.AppEnv)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	db, err := appdb.OpenPostgres(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("连接 PostgreSQL 失败", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("获取 PostgreSQL 底层连接失败", zap.Error(err))
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	jwtManager, err := auth.NewJWTManager(auth.JWTConfig{
		Secret:         cfg.JWTSecret,
		Issuer:         cfg.JWTIssuer,
		AccessTokenTTL: cfg.JWTAccessTokenTTL,
	})
	if err != nil {
		logger.Fatal("初始化 JWT Manager 失败", zap.Error(err))
	}

	userRepo := repository.NewUserRepository(db)
	workspaceRepo := repository.NewWorkspaceRepository(db)
	workflowDefinitionRepo := repository.NewWorkflowDefinitionRepository(db)
	workflowRunRepo := repository.NewWorkflowRunRepository(db)

	aiRuntimeClient, err := airuntime.NewClient(airuntime.ClientConfig{
		BaseURL: cfg.AIRuntimeURL,
	})
	if err != nil {
		logger.Fatal("初始化 AI Runtime Client 失败", zap.Error(err))
	}

	executorRegistry := workflowruntime.NewPhase4ExecutorRegistry(aiRuntimeClient)

	authService := service.NewAuthService(
		db,
		userRepo,
		workspaceRepo,
		jwtManager,
	)
	workspaceService := service.NewWorkspaceService(
		db,
		userRepo,
		workspaceRepo,
	)
	workflowRunnerService := service.NewWorkflowRunnerService(
		workflowDefinitionRepo,
		workflowRunRepo,
		workspaceService,
		executorRegistry,
	)
	workflowService := service.NewWorkflowService(
		db,
		workflowDefinitionRepo,
		workspaceService,
	)

	llmStreamService := service.NewLLMStreamService(
		aiRuntimeClient,
		workspaceService,
	)

	authHandler := handler.NewAuthHandler(authService)

	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)

	workflowRunHandler := handler.NewWorkflowRunHandler(workflowRunnerService)
	workflowHandler := handler.NewWorkflowHandler(workflowService)

	llmStreamHandler := handler.NewLLMStreamHandler(llmStreamService)

	router := server.NewRouter(
		cfg,
		logger,
		authHandler,
		workspaceHandler,
		workflowHandler,
		workflowRunHandler,
		llmStreamHandler,
		jwtManager,
	)

	httpServer := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api 服务启动", zap.String("addr", httpServer.Addr))

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("api 服务启动失败", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("api 服务开始关闭")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Fatal("api 服务关闭失败", zap.Error(err))
	}

	logger.Info("api 服务已关闭")
}

func newLogger(appEnv string) (*zap.Logger, error) {
	if appEnv == "prod" {
		return zap.NewProduction()
	}

	return zap.NewDevelopment()
}
