package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"agentflow-studio/services/api/internal/model"
	"agentflow-studio/services/api/internal/repository"
	workflowerrors "agentflow-studio/services/api/internal/workflow"
	"agentflow-studio/services/api/internal/workflowruntime"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type WorkflowService struct {
	db               *gorm.DB
	workflowRepo     *repository.WorkflowDefinitionRepository
	workspaceService *WorkspaceService
}

func NewWorkflowService(
	db *gorm.DB,
	workflowRepo *repository.WorkflowDefinitionRepository,
	workspaceService *WorkspaceService,
) *WorkflowService {
	return &WorkflowService{
		db:               db,
		workflowRepo:     workflowRepo,
		workspaceService: workspaceService,
	}
}

type CreateWorkflowInput struct {
	ActorUserID uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	Schema      *workflowruntime.WorkflowSchema
}

type GetWorkflowInput struct {
	ActorUserID uuid.UUID
	WorkspaceID uuid.UUID
	WorkflowID  uuid.UUID
}

type ListWorkflowsInput struct {
	ActorUserID uuid.UUID
	WorkspaceID uuid.UUID
}

type UpdateWorkflowInput struct {
	ActorUserID uuid.UUID
	WorkspaceID uuid.UUID
	WorkflowID  uuid.UUID
	Name        string
	Schema      *workflowruntime.WorkflowSchema
}

type WorkflowSummaryResult struct {
	ID            uuid.UUID `json:"id"`
	WorkspaceID   uuid.UUID `json:"workspace_id"`
	Name          string    `json:"name"`
	SchemaVersion string    `json:"schema_version"`
	NodeCount     int       `json:"node_count"`
	EdgeCount     int       `json:"edge_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type WorkflowDetailResult struct {
	WorkflowSummaryResult
	Schema *workflowruntime.WorkflowSchema `json:"schema"`
}

type WorkflowListResult struct {
	Items []WorkflowSummaryResult `json:"items"`
}

func (s *WorkflowService) List(
	ctx context.Context,
	input ListWorkflowsInput,
) (*WorkflowListResult, error) {
	if err := s.validateBaseInput(input.ActorUserID, input.WorkspaceID); err != nil {
		return nil, err
	}

	if _, err := s.workspaceService.RequireViewPermission(
		ctx,
		input.WorkspaceID,
		input.ActorUserID,
	); err != nil {
		return nil, err
	}

	items, err := s.workflowRepo.ListByWorkspaceID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, err
	}

	result := make([]WorkflowSummaryResult, 0, len(items))
	for _, item := range items {
		schema, err := workflowruntime.ParseWorkflowSchemaDatatypes(item.SchemaJSON)
		if err != nil {
			return nil, err
		}

		result = append(result, workflowSummaryToResult(item, schema))
	}

	return &WorkflowListResult{Items: result}, nil
}

func (s *WorkflowService) Get(
	ctx context.Context,
	input GetWorkflowInput,
) (*WorkflowDetailResult, error) {
	if err := s.validateBaseInput(input.ActorUserID, input.WorkspaceID); err != nil {
		return nil, err
	}

	if _, err := s.workspaceService.RequireViewPermission(
		ctx,
		input.WorkspaceID,
		input.ActorUserID,
	); err != nil {
		return nil, err
	}

	workflow, err := s.workflowRepo.FindByID(ctx, input.WorkspaceID, input.WorkflowID)
	if err != nil {
		return nil, mapWorkflowRepositoryError(err)
	}

	schema, err := workflowruntime.ParseWorkflowSchemaDatatypes(workflow.SchemaJSON)
	if err != nil {
		return nil, err
	}

	return &WorkflowDetailResult{
		WorkflowSummaryResult: workflowSummaryToResult(*workflow, schema),
		Schema:                schema,
	}, nil
}

func (s *WorkflowService) Create(
	ctx context.Context,
	input CreateWorkflowInput,
) (*WorkflowDetailResult, error) {
	if err := s.validateBaseInput(input.ActorUserID, input.WorkspaceID); err != nil {
		return nil, err
	}

	if _, err := s.workspaceService.RequireEditablePermission(
		ctx,
		input.WorkspaceID,
		input.ActorUserID,
	); err != nil {
		return nil, err
	}

	name, schemaJSON, schema, err := normalizeWorkflowDefinition(input.Name, input.Schema)
	if err != nil {
		return nil, err
	}

	created := &model.Workflow{
		WorkspaceID:   input.WorkspaceID,
		Name:          name,
		SchemaVersion: schema.SchemaVersion,
		SchemaJSON:    schemaJSON,
		CreatedBy:     input.ActorUserID,
		UpdatedBy:     input.ActorUserID,
	}

	if err := s.workflowRepo.Create(ctx, nil, created); err != nil {
		return nil, workflowerrors.NewError(
			workflowerrors.ErrorCodeCreateFailed,
			"Workflow 创建失败",
			err,
		)
	}

	return &WorkflowDetailResult{
		WorkflowSummaryResult: workflowSummaryToResult(*created, schema),
		Schema:                schema,
	}, nil
}

func (s *WorkflowService) Update(
	ctx context.Context,
	input UpdateWorkflowInput,
) (*WorkflowDetailResult, error) {
	if err := s.validateBaseInput(input.ActorUserID, input.WorkspaceID); err != nil {
		return nil, err
	}

	if _, err := s.workspaceService.RequireEditablePermission(
		ctx,
		input.WorkspaceID,
		input.ActorUserID,
	); err != nil {
		return nil, err
	}

	workflow, err := s.workflowRepo.FindByID(ctx, input.WorkspaceID, input.WorkflowID)
	if err != nil {
		return nil, mapWorkflowRepositoryError(err)
	}

	name, schemaJSON, schema, err := normalizeWorkflowDefinition(input.Name, input.Schema)
	if err != nil {
		return nil, err
	}

	workflow.Name = name
	workflow.SchemaVersion = schema.SchemaVersion
	workflow.SchemaJSON = schemaJSON
	workflow.UpdatedBy = input.ActorUserID
	workflow.UpdatedAt = time.Now().UTC()

	if err := s.workflowRepo.Update(ctx, nil, workflow); err != nil {
		return nil, workflowerrors.NewError(
			workflowerrors.ErrorCodeUpdateFailed,
			"Workflow 更新失败",
			err,
		)
	}

	return &WorkflowDetailResult{
		WorkflowSummaryResult: workflowSummaryToResult(*workflow, schema),
		Schema:                schema,
	}, nil
}

func (s *WorkflowService) validateBaseInput(userID, workspaceID uuid.UUID) error {
	if s.workflowRepo == nil || s.workspaceService == nil || s.db == nil {
		return workflowerrors.NewError(
			workflowerrors.ErrorCodeInvalidInput,
			"Workflow Service 未初始化",
			workflowerrors.ErrInvalidInput,
		)
	}

	if userID == uuid.Nil || workspaceID == uuid.Nil {
		return workflowerrors.NewError(
			workflowerrors.ErrorCodeInvalidInput,
			"workspace_id 和 user_id 不能为空",
			workflowerrors.ErrInvalidInput,
		)
	}

	return nil
}

func normalizeWorkflowDefinition(
	name string,
	schema *workflowruntime.WorkflowSchema,
) (string, datatypes.JSON, *workflowruntime.WorkflowSchema, error) {
	if schema == nil {
		return "", nil, nil, workflowruntime.NewRuntimeError(
			workflowruntime.ErrorCodeInvalidSchema,
			"Workflow Schema 不能为空",
			workflowruntime.ErrInvalidSchema,
		)
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = strings.TrimSpace(schema.Name)
	}
	if name == "" {
		return "", nil, nil, workflowerrors.NewError(
			workflowerrors.ErrorCodeInvalidInput,
			"Workflow 名称不能为空",
			workflowerrors.ErrInvalidInput,
		)
	}

	if len([]rune(name)) > 255 {
		return "", nil, nil, workflowerrors.NewError(
			workflowerrors.ErrorCodeInvalidInput,
			"Workflow 名称不能超过 255 个字符",
			workflowerrors.ErrInvalidInput,
		)
	}

	normalized := *schema
	normalized.Name = name
	raw, err := normalized.ToDatatypesJSON()
	if err != nil {
		return "", nil, nil, err
	}

	return name, raw, &normalized, nil
}

func workflowSummaryToResult(
	workflow model.Workflow,
	schema *workflowruntime.WorkflowSchema,
) WorkflowSummaryResult {
	return WorkflowSummaryResult{
		ID:            workflow.ID,
		WorkspaceID:   workflow.WorkspaceID,
		Name:          workflow.Name,
		SchemaVersion: workflow.SchemaVersion,
		NodeCount:     schema.Summary.NodeCount,
		EdgeCount:     schema.Summary.EdgeCount,
		CreatedAt:     workflow.CreatedAt,
		UpdatedAt:     workflow.UpdatedAt,
	}
}

func mapWorkflowRepositoryError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return workflowerrors.NewError(
			workflowerrors.ErrorCodeNotFound,
			"Workflow 不存在",
			workflowerrors.ErrNotFound,
		)
	}

	return err
}
