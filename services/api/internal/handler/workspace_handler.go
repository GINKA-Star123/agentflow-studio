package handler

import (
	"agentflow-studio/services/api/internal/requestctx"
	"agentflow-studio/services/api/internal/response"
	"agentflow-studio/services/api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WorkspaceHandler struct {
	workspaceService *service.WorkspaceService
}

func NewWorkspaceHandler(workspaceService *service.WorkspaceService) *WorkspaceHandler {
	return &WorkspaceHandler{workspaceService: workspaceService}
}

type CreateWorkspaceRequest struct {
	Name string `json:"name" binding:"required"`
}

type AddWorkspaceMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=member viewer"`
}

func (h *WorkspaceHandler) AddMember(c *gin.Context) {
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

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	var req AddWorkspaceMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BindError(c, err)
		return
	}

	result, err := h.workspaceService.AddMember(
		c.Request.Context(),
		service.AddWorkspaceMemberInput{
			ActorUserID: currentUser.ID,
			WorkspaceID: workspaceID,
			Email:       req.Email,
			Role:        req.Role,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, result)
}

func (h *WorkspaceHandler) List(c *gin.Context) {
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

	result, err := h.workspaceService.ListWorkspaces(
		c.Request.Context(),
		currentUser.ID,
	)

	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *WorkspaceHandler) Create(c *gin.Context) {
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

	var req CreateWorkspaceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BindError(c, err)
		return
	}

	result, err := h.workspaceService.CreateWorkspace(
		c.Request.Context(),
		service.CreateWorkspaceInput{
			UserID: currentUser.ID,
			Name:   req.Name,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, result)
}

type UpdateWorkspaceMemberRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=admin member viewer"`
}

func (h *WorkspaceHandler) ListMembers(c *gin.Context) {
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

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	result, err := h.workspaceService.ListMembers(
		c.Request.Context(),
		service.ListWorkspaceMembersInput{
			ActorUserID: currentUser.ID,
			WorkspaceID: workspaceID,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *WorkspaceHandler) UpdateMemberRole(c *gin.Context) {
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

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	targetUserID, ok := parseUUIDParam(c, "user_id")
	if !ok {
		return
	}

	var req UpdateWorkspaceMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BindError(c, err)
		return
	}

	result, err := h.workspaceService.UpdateMemberRole(
		c.Request.Context(),
		service.UpdateWorkspaceMemberRoleInput{
			ActorUserID:  currentUser.ID,
			WorkspaceID:  workspaceID,
			TargetUserID: targetUserID,
			Role:         req.Role,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *WorkspaceHandler) RemoveMember(c *gin.Context) {
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

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	targetUserID, ok := parseUUIDParam(c, "user_id")
	if !ok {
		return
	}

	if err := h.workspaceService.RemoveMember(
		c.Request.Context(),
		service.RemoveWorkspaceMemberInput{
			ActorUserID:  currentUser.ID,
			WorkspaceID:  workspaceID,
			TargetUserID: targetUserID,
		},
	); err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, gin.H{
		"removed": true,
	})
}

func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, bool) {
	value := c.Param(name)

	id, err := uuid.Parse(value)
	if err != nil {
		response.Fail(
			c,
			400,
			"INVALID_ARGUMENT",
			"路径参数无效",
			gin.H{
				"param": name,
				"value": value,
			},
		)
		return uuid.Nil, false
	}

	return id, true
}
