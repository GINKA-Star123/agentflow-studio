package repository

import (
	"context"

	"agentflow-studio/services/api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkflowRunRepository struct {
	db *gorm.DB
}

func NewWorkflowRunRepository(db *gorm.DB) *WorkflowRunRepository {
	return &WorkflowRunRepository{
		db: db,
	}
}

func (r *WorkflowRunRepository) CreateRun(
	ctx context.Context,
	tx *gorm.DB,
	run *model.WorkflowRun,
) error {
	db := r.useDB(tx)

	return db.WithContext(ctx).Create(run).Error
}

func (r *WorkflowRunRepository) FindRunByID(
	ctx context.Context,
	workspaceID uuid.UUID,
	runID uuid.UUID,
) (*model.WorkflowRun, error) {
	var run model.WorkflowRun

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND id = ?", workspaceID, runID).
		First(&run).
		Error

	if err != nil {
		return nil, err
	}

	return &run, nil
}

func (r *WorkflowRunRepository) ListRunsByWorkflow(
	ctx context.Context,
	workspaceID uuid.UUID,
	workflowID uuid.UUID,
	limit int,
) ([]model.WorkflowRun, error) {
	if limit <= 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	var runs []model.WorkflowRun

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND workflow_id = ?", workspaceID, workflowID).
		Order("created_at DESC").
		Limit(limit).
		Find(&runs).
		Error

	if err != nil {
		return nil, err
	}

	return runs, nil
}

func (r *WorkflowRunRepository) UpdateRun(
	ctx context.Context,
	tx *gorm.DB,
	run *model.WorkflowRun,
) error {
	db := r.useDB(tx)

	result := db.WithContext(ctx).
		Model(&model.WorkflowRun{}).
		Where("workspace_id = ? AND id = ?", run.WorkspaceID, run.ID).
		Updates(map[string]any{
			"status":      run.Status,
			"output":      run.Output,
			"error":       run.Error,
			"started_at":  run.StartedAt,
			"finished_at": run.FinishedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *WorkflowRunRepository) CreateNodeExecution(
	ctx context.Context,
	tx *gorm.DB,
	nodeExecution *model.NodeExecution,
) error {
	db := r.useDB(tx)

	return db.WithContext(ctx).Create(nodeExecution).Error
}

func (r *WorkflowRunRepository) FindNodeExecutionByID(
	ctx context.Context,
	workspaceID uuid.UUID,
	nodeExecutionID uuid.UUID,
) (*model.NodeExecution, error) {
	var nodeExecution model.NodeExecution

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND id = ?", workspaceID, nodeExecutionID).
		First(&nodeExecution).
		Error

	if err != nil {
		return nil, err
	}

	return &nodeExecution, nil
}

func (r *WorkflowRunRepository) ListNodeExecutions(
	ctx context.Context,
	workspaceID uuid.UUID,
	runID uuid.UUID,
) ([]model.NodeExecution, error) {
	var nodeExecutions []model.NodeExecution

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND run_id = ?", workspaceID, runID).
		Order("created_at ASC").
		Order("sequence ASC").
		Find(&nodeExecutions).
		Error

	if err != nil {
		return nil, err
	}

	return nodeExecutions, nil
}

func (r *WorkflowRunRepository) UpdateNodeExecution(
	ctx context.Context,
	tx *gorm.DB,
	nodeExecution *model.NodeExecution,
) error {
	db := r.useDB(tx)

	result := db.WithContext(ctx).
		Model(&model.NodeExecution{}).
		Where("workspace_id = ? AND id = ?", nodeExecution.WorkspaceID, nodeExecution.ID).
		Updates(map[string]any{
			"status":      nodeExecution.Status,
			"input":       nodeExecution.Input,
			"output":      nodeExecution.Output,
			"error":       nodeExecution.Error,
			"token_usage": nodeExecution.TokenUsage,
			"latency_ms":  nodeExecution.LatencyMS,
			"started_at":  nodeExecution.StartedAt,
			"finished_at": nodeExecution.FinishedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *WorkflowRunRepository) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}

	return r.db
}
