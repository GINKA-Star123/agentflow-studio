package handler

import (
	"agentflow-studio/services/api/internal/response"
	"agentflow-studio/services/api/internal/service"
	"agentflow-studio/services/api/internal/workflowruntime"

	"github.com/gin-gonic/gin"
)

type WorkflowHandler struct {
	workflowService *service.WorkflowService
}

func NewWorkflowHandler(workflowService *service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{workflowService: workflowService}
}

type CreateWorkflowRequest struct {
	Name   string                          `json:"name"`
	Schema *workflowruntime.WorkflowSchema `json:"schema"`
}

type UpdateWorkflowRequest struct {
	Name   string                          `json:"name"`
	Schema *workflowruntime.WorkflowSchema `json:"schema"`
}

func (h *WorkflowHandler) List(c *gin.Context) {
	currentUser, ok := getCurrentUserOrAbort(c)
	if !ok {
		return
	}

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	result, err := h.workflowService.List(
		c.Request.Context(),
		service.ListWorkflowsInput{
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

func (h *WorkflowHandler) Create(c *gin.Context) {
	currentUser, ok := getCurrentUserOrAbort(c)
	if !ok {
		return
	}

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BindError(c, err)
		return
	}

	result, err := h.workflowService.Create(
		c.Request.Context(),
		service.CreateWorkflowInput{
			ActorUserID: currentUser.ID,
			WorkspaceID: workspaceID,
			Name:        req.Name,
			Schema:      req.Schema,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, result)
}

func (h *WorkflowHandler) Get(c *gin.Context) {
	currentUser, ok := getCurrentUserOrAbort(c)
	if !ok {
		return
	}

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	workflowID, ok := parseUUIDParam(c, "workflow_id")
	if !ok {
		return
	}

	result, err := h.workflowService.Get(
		c.Request.Context(),
		service.GetWorkflowInput{
			ActorUserID: currentUser.ID,
			WorkspaceID: workspaceID,
			WorkflowID:  workflowID,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *WorkflowHandler) Update(c *gin.Context) {
	currentUser, ok := getCurrentUserOrAbort(c)
	if !ok {
		return
	}

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	workflowID, ok := parseUUIDParam(c, "workflow_id")
	if !ok {
		return
	}

	var req UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BindError(c, err)
		return
	}

	result, err := h.workflowService.Update(
		c.Request.Context(),
		service.UpdateWorkflowInput{
			ActorUserID: currentUser.ID,
			WorkspaceID: workspaceID,
			WorkflowID:  workflowID,
			Name:        req.Name,
			Schema:      req.Schema,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, result)
}
