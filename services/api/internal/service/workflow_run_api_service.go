package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"agentflow-studio/services/api/internal/model"
	"agentflow-studio/services/api/internal/workflowruntime"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GetWorkflowRunInput struct {
	ActorUserID uuid.UUID
	WorkspaceID uuid.UUID
	RunID       uuid.UUID
}

type ListWorkflowRunNodesInput struct {
	ActorUserID uuid.UUID
	WorkspaceID uuid.UUID
	RunID       uuid.UUID
}

type CancelWorkflowRunInput struct {
	ActorUserID uuid.UUID
	WorkspaceID uuid.UUID
	RunID       uuid.UUID
}

type WorkflowRunDetailResult struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	TriggeredBy uuid.UUID `json:"triggered_by"`
	Status      string    `json:"status"`

	Input  any `json:"input,omitempty"`
	Output any `json:"output,omitempty"`
	Error  any `json:"error,omitempty"`

	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type WorkflowRunNodeExecutionResult struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	RunID       uuid.UUID `json:"run_id"`

	NodeID   string `json:"node_id"`
	NodeType string `json:"node_type"`
	Sequence int    `json:"sequence"`
	Status   string `json:"status"`

	Input      any   `json:"input,omitempty"`
	Output     any   `json:"output,omitempty"`
	Error      any   `json:"error,omitempty"`
	TokenUsage any   `json:"token_usage,omitempty"`
	LatencyMS  int64 `json:"latency_ms"`

	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type WorkflowRunNodeExecutionListResult struct {
	Items []WorkflowRunNodeExecutionResult `json:"items"`
}

func (s *WorkflowRunnerService) GetRun(
	ctx context.Context,
	input GetWorkflowRunInput,
) (*WorkflowRunDetailResult, error) {
	if err := s.requireRunViewPermission(ctx, input.WorkspaceID, input.ActorUserID); err != nil {
		return nil, err
	}

	run, err := s.findRun(ctx, input.WorkspaceID, input.RunID)
	if err != nil {
		return nil, err
	}

	return workflowRunToDetailResult(run), nil
}

func (s *WorkflowRunnerService) ListRunNodes(
	ctx context.Context,
	input ListWorkflowRunNodesInput,
) (*WorkflowRunNodeExecutionListResult, error) {
	if err := s.requireRunViewPermission(ctx, input.WorkspaceID, input.ActorUserID); err != nil {
		return nil, err
	}

	if _, err := s.findRun(ctx, input.WorkspaceID, input.RunID); err != nil {
		return nil, err
	}

	nodeExecutions, err := s.runRepo.ListNodeExecutions(ctx, input.WorkspaceID, input.RunID)
	if err != nil {
		return nil, err
	}

	items := make([]WorkflowRunNodeExecutionResult, 0, len(nodeExecutions))
	for _, item := range nodeExecutions {
		items = append(items, nodeExecutionToResult(item))
	}

	return &WorkflowRunNodeExecutionListResult{
		Items: items,
	}, nil
}

func (s *WorkflowRunnerService) CancelRun(
	ctx context.Context,
	input CancelWorkflowRunInput,
) (*WorkflowRunDetailResult, error) {
	if err := s.requireRunPermission(ctx, input.WorkspaceID, input.ActorUserID); err != nil {
		return nil, err
	}

	run, err := s.findRun(ctx, input.WorkspaceID, input.RunID)
	if err != nil {
		return nil, err
	}

	if run.IsTerminal() {
		return nil, workflowruntime.NewRuntimeErrorWithDetails(
			workflowruntime.ErrorCodeRunAlreadyTerminal,
			"Workflow Run 已经结束，不能取消",
			workflowruntime.ErrRunAlreadyTerminal,
			map[string]any{
				"run_id": run.ID.String(),
				"status": string(run.Status),
			},
		)
	}

	finishedAt := time.Now().UTC()
	run.MarkCanceled(
		finishedAt,
		toDatatypesJSON(workflowruntime.JSONMap{
			"code":    "RUNTIME_CANCELED",
			"message": "Workflow Run 已取消",
		}),
	)

	if err := s.runRepo.UpdateRun(ctx, nil, run); err != nil {
		return nil, workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeCancelFailed,
			"Workflow Run 取消失败",
			err,
		)
	}

	return workflowRunToDetailResult(run), nil
}

func (s *WorkflowRunnerService) requireRunViewPermission(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) error {
	if s.workspaceService == nil {
		return workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeInvalidInput,
			"WorkspaceService 未初始化",
			workflowruntime.ErrInvalidInput,
		)
	}

	_, err := s.workspaceService.RequireViewPermission(ctx, workspaceID, userID)
	return err
}

func (s *WorkflowRunnerService) findRun(
	ctx context.Context,
	workspaceID uuid.UUID,
	runID uuid.UUID,
) (*model.WorkflowRun, error) {
	if runID == uuid.Nil {
		return nil, workflowruntime.NewRuntimeErrorWithDetails(
			workflowruntime.ErrorCodeInvalidInput,
			"run_id 不能为空",
			workflowruntime.ErrInvalidInput,
			map[string]any{
				"field": "run_id",
			},
		)
	}

	run, err := s.runRepo.FindRunByID(ctx, workspaceID, runID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, workflowruntime.NewRuntimeError(
				workflowruntime.ErrorCodeRunNotFound,
				"Workflow Run 不存在",
				workflowruntime.ErrRunNotFound,
			)
		}

		return nil, err
	}

	return run, nil
}

func workflowRunToDetailResult(run *model.WorkflowRun) *WorkflowRunDetailResult {
	return &WorkflowRunDetailResult{
		ID:          run.ID,
		WorkspaceID: run.WorkspaceID,
		WorkflowID:  run.WorkflowID,
		TriggeredBy: run.TriggeredBy,
		Status:      string(run.Status),
		Input:       decodeDatatypesJSON(run.Input),
		Output:      decodeDatatypesJSON(run.Output),
		Error:       decodeDatatypesJSON(run.Error),
		StartedAt:   run.StartedAt,
		FinishedAt:  run.FinishedAt,
		CreatedAt:   run.CreatedAt,
		UpdatedAt:   run.UpdatedAt,
	}
}

func nodeExecutionToResult(
	nodeExecution model.NodeExecution,
) WorkflowRunNodeExecutionResult {
	return WorkflowRunNodeExecutionResult{
		ID:          nodeExecution.ID,
		WorkspaceID: nodeExecution.WorkspaceID,
		RunID:       nodeExecution.RunID,
		NodeID:      nodeExecution.NodeID,
		NodeType:    nodeExecution.NodeType,
		Sequence:    nodeExecution.Sequence,
		Status:      string(nodeExecution.Status),
		Input:       decodeDatatypesJSON(nodeExecution.Input),
		Output:      decodeDatatypesJSON(nodeExecution.Output),
		Error:       decodeDatatypesJSON(nodeExecution.Error),
		TokenUsage:  decodeDatatypesJSON(nodeExecution.TokenUsage),
		LatencyMS:   nodeExecution.LatencyMS,
		StartedAt:   nodeExecution.StartedAt,
		FinishedAt:  nodeExecution.FinishedAt,
		CreatedAt:   nodeExecution.CreatedAt,
		UpdatedAt:   nodeExecution.UpdatedAt,
	}
}

func decodeDatatypesJSON(raw datatypes.JSON) any {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil
	}

	var value any
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return string(trimmed)
	}

	return value
}
