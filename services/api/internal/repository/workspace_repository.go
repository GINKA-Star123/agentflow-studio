package repository

import (
	"context"

	"agentflow-studio/services/api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkspaceRepository struct {
	db *gorm.DB
}

func NewWorkspaceRepository(db *gorm.DB) *WorkspaceRepository {
	return &WorkspaceRepository{
		db: db,
	}
}

func (r *WorkspaceRepository) Create(ctx context.Context, tx *gorm.DB, workspace *model.Workspace) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	return db.WithContext(ctx).Create(workspace).Error
}

func (r *WorkspaceRepository) CreateMember(ctx context.Context, tx *gorm.DB, member *model.WorkspaceMember) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	return db.WithContext(ctx).Create(member).Error
}

func (r *WorkspaceRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Workspace, error) {
	var workspace model.Workspace

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&workspace).
		Error

	if err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (r *WorkspaceRepository) FindMember(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) (*model.WorkspaceMember, error) {
	var member model.WorkspaceMember

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		First(&member).
		Error

	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *WorkspaceRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]model.Workspace, error) {
	var workspaces []model.Workspace

	err := r.db.WithContext(ctx).
		Joins("JOIN workspace_members ON workspace_members.workspace_id = workspaces.id").
		Where("workspace_members.user_id = ?", userID).
		Order("workspaces.created_at ASC").
		Find(&workspaces).
		Error

	if err != nil {
		return nil, err
	}

	return workspaces, nil
}

func (r *WorkspaceRepository) ListMembers(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]model.WorkspaceMember, error) {
	var members []model.WorkspaceMember

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("workspace_id = ?", workspaceID).
		Order("created_at ASC").
		Find(&members).
		Error

	if err != nil {
		return nil, err
	}

	return members, nil

}

func (r *WorkspaceRepository) FindMemberWithUser(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) (*model.WorkspaceMember, error) {
	var member model.WorkspaceMember

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		First(&member).
		Error

	if err != nil {
		return nil, err
	}

	return &member, nil
}

func (r *WorkspaceRepository) UpdateMemberRole(
	ctx context.Context,
	tx *gorm.DB,
	workspaceID uuid.UUID,
	userID uuid.UUID,
	role model.WorkspaceRole,
) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	result := db.WithContext(ctx).
		Model(&model.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Update("role", role)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *WorkspaceRepository) DeleteMember(
	ctx context.Context,
	tx *gorm.DB,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) error {
	db := r.db
	if tx != nil {
		db = tx
	}

	result := db.WithContext(ctx).
		Where("workspace_id = ? AND user_id = ?", workspaceID, userID).
		Delete(&model.WorkspaceMember{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *WorkspaceRepository) CountMemberByRole(
	ctx context.Context,
	tx *gorm.DB,
	workspaceID uuid.UUID,
	role model.WorkspaceRole,
) (int64, error) {
	db := r.db
	if tx != nil {
		db = tx
	}

	var count int64

	err := db.WithContext(ctx).
		Model(&model.WorkspaceMember{}).
		Where("workspace_id = ? AND role = ?", workspaceID, role).
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}
