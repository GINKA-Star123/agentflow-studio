package handler

import (
	"errors"
	"io"

	"agentflow-studio/services/api/internal/requestctx"
	"agentflow-studio/services/api/internal/response"
	"agentflow-studio/services/api/internal/service"
	"agentflow-studio/services/api/internal/workflowruntime"

	"github.com/gin-gonic/gin"
)

type WorkflowRunHandler struct {
	workflowRunnerService *service.WorkflowRunnerService
}

func NewWorkflowRunHandler(
	workflowRunnerService *service.WorkflowRunnerService,
) *WorkflowRunHandler {
	return &WorkflowRunHandler{
		workflowRunnerService: workflowRunnerService,
	}
}

type StartWorkflowRunRequest struct {
	Input   workflowruntime.JSONMap `json:"input"`
	TraceID string                  `json:"trace_id"`
}

func (h *WorkflowRunHandler) Start(c *gin.Context) {
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

	var req StartWorkflowRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if !errors.Is(err, io.EOF) {
			response.BindError(c, err)
			return
		}
	}

	if req.Input == nil {
		req.Input = workflowruntime.JSONMap{}
	}

	result, err := h.workflowRunnerService.RunWorkflow(
		c.Request.Context(),
		service.StartWorkflowRunInput{
			WorkspaceID: workspaceID,
			WorkflowID:  workflowID,
			UserID:      currentUser.ID,
			Input:       req.Input,
			TraceID:     req.TraceID,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Created(c, result)
}

func (h *WorkflowRunHandler) Get(c *gin.Context) {
	currentUser, ok := getCurrentUserOrAbort(c)
	if !ok {
		return
	}

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	runID, ok := parseUUIDParam(c, "run_id")
	if !ok {
		return
	}

	result, err := h.workflowRunnerService.GetRun(
		c.Request.Context(),
		service.GetWorkflowRunInput{
			ActorUserID: currentUser.ID,
			WorkspaceID: workspaceID,
			RunID:       runID,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *WorkflowRunHandler) ListNodes(c *gin.Context) {
	currentUser, ok := getCurrentUserOrAbort(c)
	if !ok {
		return
	}

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	runID, ok := parseUUIDParam(c, "run_id")
	if !ok {
		return
	}

	result, err := h.workflowRunnerService.ListRunNodes(
		c.Request.Context(),
		service.ListWorkflowRunNodesInput{
			ActorUserID: currentUser.ID,
			WorkspaceID: workspaceID,
			RunID:       runID,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, result)
}

func (h *WorkflowRunHandler) Cancel(c *gin.Context) {
	currentUser, ok := getCurrentUserOrAbort(c)
	if !ok {
		return
	}

	workspaceID, ok := parseUUIDParam(c, "workspace_id")
	if !ok {
		return
	}

	runID, ok := parseUUIDParam(c, "run_id")
	if !ok {
		return
	}

	result, err := h.workflowRunnerService.CancelRun(
		c.Request.Context(),
		service.CancelWorkflowRunInput{
			ActorUserID: currentUser.ID,
			WorkspaceID: workspaceID,
			RunID:       runID,
		},
	)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.OK(c, gin.H{
		"canceled": true,
		"run":      result,
	})
}

func getCurrentUserOrAbort(c *gin.Context) (*requestctx.CurrentUser, bool) {
	currentUser, err := requestctx.GetCurrentUser(c)
	if err != nil {
		response.Fail(
			c,
			401,
			"UNAUTHORIZED",
			"未认证",
			nil,
		)
		return nil, false
	}

	return currentUser, true
}
