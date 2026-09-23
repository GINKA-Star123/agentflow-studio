package handler

import (
	"agentflow-studio/services/api/internal/requestctx"
	"agentflow-studio/services/api/internal/response"
	"agentflow-studio/services/api/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required"`
	DisplayName   string `json:"display_name"`
	WorkspaceName string `json:"workspace_name"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		response.BindError(c, err)
		return
	}

	result, err := h.authService.Register(c.Request.Context(), service.RegisterInput{
		Email:         req.Email,
		Password:      req.Password,
		DisplayName:   req.DisplayName,
		WorkspaceName: req.WorkspaceName,
	})

	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, result)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BindError(c, err)
		return
	}

	result, err := h.authService.Login(c.Request.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		response.FromError(c, err)
	}
	response.OK(c, result)
}

func (h *AuthHandler) Me(c *gin.Context) {
	currentUser, err := requestctx.GetCurrentUser(c)
	if err != nil {
		response.Fail(
			c,
			401,
			"UNAUTHORIZED",
			"未认证",
			nil,
		)
		return
	}

	result, err := h.authService.GetCurrentUser(c.Request.Context(), currentUser.ID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, result)
}
