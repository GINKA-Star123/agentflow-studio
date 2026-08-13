package repository

import (
	"context"

	"agentflow-studio/services/api/internal/model"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type RunnableWorkflowDefinition struct {
	ID          uuid.UUID      `gorm:"column:id"`
	WorkspaceID uuid.UUID      `gorm:"column:workspace_id"`
	Name        string         `gorm:"column:name"`
	SchemaJSON  datatypes.JSON `gorm:"column:schema_json"`
}

type WorkflowDefinitionRepository struct {
	db *gorm.DB
}

func NewWorkflowDefinitionRepository(db *gorm.DB) *WorkflowDefinitionRepository {
	return &WorkflowDefinitionRepository{
		db: db,
	}
}

func (r *WorkflowDefinitionRepository) ListByWorkspaceID(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]model.Workflow, error) {
	var workflows []model.Workflow

	err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("updated_at DESC").
		Find(&workflows).
		Error

	return workflows, err
}

func (r *WorkflowDefinitionRepository) FindByID(
	ctx context.Context,
	workspaceID uuid.UUID,
	workflowID uuid.UUID,
) (*model.Workflow, error) {
	var workflow model.Workflow

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND id = ?", workspaceID, workflowID).
		First(&workflow).
		Error

	if err != nil {
		return nil, err
	}

	return &workflow, nil
}

func (r *WorkflowDefinitionRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	workflow *model.Workflow,
) error {
	return r.useDB(tx).WithContext(ctx).Create(workflow).Error
}

func (r *WorkflowDefinitionRepository) Update(
	ctx context.Context,
	tx *gorm.DB,
	workflow *model.Workflow,
) error {
	result := r.useDB(tx).WithContext(ctx).
		Model(&model.Workflow{}).
		Where("workspace_id = ? AND id = ?", workflow.WorkspaceID, workflow.ID).
		Updates(map[string]any{
			"name":           workflow.Name,
			"schema_version": workflow.SchemaVersion,
			"schema_json":    workflow.SchemaJSON,
			"updated_by":     workflow.UpdatedBy,
			"updated_at":     workflow.UpdatedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *WorkflowDefinitionRepository) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}

	return r.db
}

func (r *WorkflowDefinitionRepository) FindRunnableByID(
	ctx context.Context,
	workspaceID uuid.UUID,
	workflowID uuid.UUID,
) (*RunnableWorkflowDefinition, error) {
	var definition RunnableWorkflowDefinition

	err := r.db.WithContext(ctx).
		Table("workflows").
		Select("id, workspace_id, name, schema_json").
		Where("workspace_id = ? AND id = ? AND deleted_at IS NULL", workspaceID, workflowID).
		Take(&definition).
		Error

	if err != nil {
		return nil, err
	}

	return &definition, nil
}
