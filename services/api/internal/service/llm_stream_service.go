package service

import (
	"agentflow-studio/services/api/internal/airuntime"
	"agentflow-studio/services/api/internal/model"
	"agentflow-studio/services/api/internal/workflowruntime"
	"agentflow-studio/services/api/internal/workspace"
	"context"

	"github.com/google/uuid"
)

type LLMStreamService struct {
	aiRuntimeClient  *airuntime.Client
	workspaceService *WorkspaceService
}

func NewLLMStreamService(
	aiRuntimeClient *airuntime.Client,
	workspaceService *WorkspaceService,
) *LLMStreamService {
	return &LLMStreamService{
		aiRuntimeClient:  aiRuntimeClient,
		workspaceService: workspaceService,
	}
}

type StartLLMStreamInput struct {
	ActorUserID uuid.UUID
	WorkspaceID uuid.UUID
	Request     airuntime.ChatRequest
}

func (s *LLMStreamService) StreamChat(
	ctx context.Context,
	input StartLLMStreamInput,
) (*airuntime.ChatStream, error) {
	if s.aiRuntimeClient == nil {
		return nil, workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeAIRuntimeError,
			"AI Runtime Client 未初始化",
			workflowruntime.ErrAIRuntimeError,
		)
	}

	if s.workspaceService == nil {
		return nil, workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeInvalidInput,
			"Workspace Service 未初始化",
			workflowruntime.ErrInvalidInput,
		)
	}

	member, err := s.workspaceService.RequireMember(
		ctx,
		input.WorkspaceID,
		input.ActorUserID,
	)
	if err != nil {
		return nil, err
	}

	if member.Role == model.WorkspaceRoleViewer {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodePermissionDenied,
			"Viewer 无权发起LLM streaming",
			workspace.ErrPermissionDenied,
		)
	}

	if err := input.Request.Validate(); err != nil {
		return nil, workflowruntime.NewRuntimeErrorWithDetails(
			workflowruntime.ErrorCodeInvalidInput,
			"LLM Streaming 请求参数无效",
			err,
			map[string]any{
				"reason": err.Error(),
			},
		)
	}

	input.Request.Metadata = enrichLLMStreamMetadata(
		input.Request.Metadata,
		input.WorkspaceID,
		input.ActorUserID,
	)

	stream, err := s.aiRuntimeClient.StreamChat(ctx, input.Request)
	if err != nil {
		return nil, workflowruntime.NewRuntimeErrorWithDetails(
			workflowruntime.ErrorCodeAIRuntimeError,
			"AI Runtime Streaming 调用失败",
			err,
			map[string]any{
				"provider": input.Request.Provider,
				"model":    input.Request.Model,
				"reason":   err.Error(),
			},
		)
	}

	return stream, nil
}

func enrichLLMStreamMetadata(
	metadata map[string]any,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) map[string]any {
	result := map[string]any{}

	for key, value := range metadata {
		result[key] = value
	}

	result["workspace_id"] = workspaceID.String()
	result["user_id"] = userID.String()

	return result
}
