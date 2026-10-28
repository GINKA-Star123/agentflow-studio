package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"agentflow-studio/services/api/internal/model"
	"agentflow-studio/services/api/internal/repository"
	"agentflow-studio/services/api/internal/workspace"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkspaceService struct {
	db            *gorm.DB
	userRepo      *repository.UserRepository
	workspaceRepo *repository.WorkspaceRepository
}

func NewWorkspaceService(
	db *gorm.DB,
	userRepo *repository.UserRepository,
	workspaceRepo *repository.WorkspaceRepository,
) *WorkspaceService {
	return &WorkspaceService{
		db:            db,
		userRepo:      userRepo,
		workspaceRepo: workspaceRepo,
	}
}

type CreateWorkspaceInput struct {
	UserID uuid.UUID
	Name   string
}

type WorkspaceResult struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	OwnerID uuid.UUID `json:"owner_id"`
	Role    string    `json:"role"`
}

type WorkspaceListResult struct {
	Items []WorkspaceResult `json:"items"`
}

type ListWorkspaceMembersInput struct {
	ActorUserID uuid.UUID
	WorkspaceID uuid.UUID
}

type UpdateWorkspaceMemberRoleInput struct {
	ActorUserID  uuid.UUID
	WorkspaceID  uuid.UUID
	TargetUserID uuid.UUID
	Role         string
}

type RemoveWorkspaceMemberInput struct {
	ActorUserID  uuid.UUID
	WorkspaceID  uuid.UUID
	TargetUserID uuid.UUID
}

type WorkspaceMemberResult struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
}

type WorkspaceMemberListResult struct {
	Items []WorkspaceMemberResult `json:"items"`
}

type AddWorkspaceMemberInput struct {
	ActorUserID uuid.UUID
	WorkspaceID uuid.UUID
	Email       string
	Role        string
}

func (s *WorkspaceService) AddMember(
	ctx context.Context,
	input AddWorkspaceMemberInput,
) (*WorkspaceMemberResult, error) {
	actorMember, err := s.RequireManagePermission(ctx, input.WorkspaceID, input.ActorUserID)
	if err != nil {
		return nil, err
	}

	targetRole, err := parseAddableWorkspaceRole(input.Role)
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodeInvalidInput,
			"用户邮箱不能为空",
			workspace.ErrInvalidInput,
		)
	}

	targetUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, workspace.NewWorkspaceError(
				workspace.ErrorCodeUserNotFound,
				"用户不存在",
				workspace.ErrUserNotFound,
			)
		}

		return nil, err
	}

	if targetUser.ID == input.ActorUserID {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodeInvalidInput,
			"不能把自己重复添加到 Workspace",
			workspace.ErrInvalidInput,
		)
	}

	if err := canAddWorkspaceMember(actorMember, targetRole); err != nil {
		return nil, err
	}

	_, err = s.workspaceRepo.FindMember(ctx, input.WorkspaceID, targetUser.ID)
	if err == nil {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodeMemberAlreadyExists,
			"用户已经是该 Workspace 成员",
			workspace.ErrMemberAlreadyExists,
		)
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	newMember := &model.WorkspaceMember{
		WorkspaceID: input.WorkspaceID,
		UserID:      targetUser.ID,
		Role:        targetRole,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.workspaceRepo.CreateMember(ctx, tx, newMember)
	})
	if err != nil {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodeCreateFailed,
			"Workspace 成员添加失败",
			err,
		)
	}

	createdMember, err := s.workspaceRepo.FindMemberWithUser(
		ctx,
		input.WorkspaceID,
		targetUser.ID,
	)
	if err != nil {
		return nil, err
	}

	return ptr(workspaceMemberToResult(*createdMember)), nil
}

func parseAddableWorkspaceRole(role string) (model.WorkspaceRole, error) {
	normalized := strings.ToLower(strings.TrimSpace(role))

	switch model.WorkspaceRole(normalized) {
	case model.WorkspaceRoleMember,
		model.WorkspaceRoleViewer:
		return model.WorkspaceRole(normalized), nil

	default:
		return "", workspace.NewWorkspaceError(
			workspace.ErrorCodeInvalidRole,
			"只能添加 member 或 viewer 角色",
			workspace.ErrInvalidRole,
		)
	}
}

func canAddWorkspaceMember(
	actor *model.WorkspaceMember,
	targetRole model.WorkspaceRole,
) error {
	if actor.Role == model.WorkspaceRoleOwner {
		return nil
	}

	if actor.Role == model.WorkspaceRoleAdmin {
		if targetRole == model.WorkspaceRoleMember || targetRole == model.WorkspaceRoleViewer {
			return nil
		}
	}

	return workspace.NewWorkspaceError(
		workspace.ErrorCodePermissionDenied,
		"当前用户没有添加成员的权限",
		workspace.ErrPermissionDenied,
	)
}
func (s *WorkspaceService) ListWorkspaces(
	ctx context.Context,
	userID uuid.UUID,
) (*WorkspaceListResult, error) {
	if userID == uuid.Nil {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodeForbidden,
			"当前用户无效",
			workspace.ErrForbidden,
		)
	}

	workspaces, err := s.workspaceRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]WorkspaceResult, 0, len(workspaces))

	for _, item := range workspaces {
		member, err := s.workspaceRepo.FindMember(ctx, item.ID, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}

			return nil, err
		}

		items = append(items, WorkspaceResult{
			ID:      item.ID,
			Name:    item.Name,
			OwnerID: item.OwnerID,
			Role:    string(member.Role),
		})
	}

	return &WorkspaceListResult{
		Items: items,
	}, nil
}

func (s *WorkspaceService) CreateWorkspace(
	ctx context.Context,
	input CreateWorkspaceInput,
) (*WorkspaceResult, error) {
	if input.UserID == uuid.Nil {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodeForbidden,
			"当前用户无效",
			workspace.ErrForbidden,
		)
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodeInvalidInput,
			"Workspace 名称不能为空",
			workspace.ErrInvalidInput,
		)
	}

	if len([]rune(name)) > 100 {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodeInvalidInput,
			"Workspace 名称不能超过 100 个字符",
			workspace.ErrInvalidInput,
		)
	}

	var createdWorkspace *model.Workspace
	var createdMember *model.WorkspaceMember

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		newWorkspace := &model.Workspace{
			Name:    name,
			OwnerID: input.UserID,
		}

		if err := s.workspaceRepo.Create(ctx, tx, newWorkspace); err != nil {
			return workspace.NewWorkspaceError(
				workspace.ErrorCodeCreateFailed,
				"Workspace 创建失败",
				err,
			)
		}

		newMember := &model.WorkspaceMember{
			WorkspaceID: newWorkspace.ID,
			UserID:      input.UserID,
			Role:        model.WorkspaceRoleOwner,
		}

		if err := s.workspaceRepo.CreateMember(ctx, tx, newMember); err != nil {
			return workspace.NewWorkspaceError(
				workspace.ErrorCodeCreateFailed,
				"Workspace 成员关系创建失败",
				err,
			)
		}

		createdWorkspace = newWorkspace
		createdMember = newMember

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &WorkspaceResult{
		ID:      createdWorkspace.ID,
		Name:    createdWorkspace.Name,
		OwnerID: createdWorkspace.OwnerID,
		Role:    string(createdMember.Role),
	}, nil
}

func (s *WorkspaceService) RequireMember(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) (*model.WorkspaceMember, error) {
	if workspaceID == uuid.Nil {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodeInvalidInput,
			"Workspace ID 无效",
			workspace.ErrInvalidInput,
		)
	}

	if userID == uuid.Nil {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodeForbidden,
			"当前用户无效",
			workspace.ErrForbidden,
		)
	}

	_, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, workspace.NewWorkspaceError(
				workspace.ErrorCodeNotFound,
				"Workspace 不存在",
				workspace.ErrNotFound,
			)
		}

		return nil, err
	}

	member, err := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, workspace.NewWorkspaceError(
				workspace.ErrorCodeMemberNotFound,
				"当前用户不是该 Workspace 成员",
				workspace.ErrMemberNotFound,
			)
		}

		return nil, err
	}

	return member, nil
}

func (s *WorkspaceService) RequireManagePermission(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) (*model.WorkspaceMember, error) {
	member, err := s.RequireMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}

	if !member.CanManageWorkspace() {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodePermissionDenied,
			"当前用户没有管理该 Workspace 的权限",
			workspace.ErrPermissionDenied,
		)
	}

	return member, nil
}

func (s *WorkspaceService) RequireViewPermission(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) (*model.WorkspaceMember, error) {
	member, err := s.RequireMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}

	if !member.CanViewWorkspace() {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodePermissionDenied,
			"当前用户没有查看该 Workspace 的权限",
			workspace.ErrPermissionDenied,
		)
	}

	return member, nil
}

// RequireEditablePermission 允许 Workflow 等业务资源由 member 及以上角色编辑。
// viewer 保留只读权限，不允许创建或更新 Workflow。
func (s *WorkspaceService) RequireEditablePermission(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) (*model.WorkspaceMember, error) {
	member, err := s.RequireMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}

	if member.Role == model.WorkspaceRoleViewer {
		return nil, workspace.NewWorkspaceError(
			workspace.ErrorCodePermissionDenied,
			"viewer 没有编辑 Workflow 的权限",
			workspace.ErrPermissionDenied,
		)
	}

	return member, nil
}

func (s *WorkspaceService) ListMembers(
	ctx context.Context,
	input ListWorkspaceMembersInput,
) (*WorkspaceMemberListResult, error) {
	if _, err := s.RequireViewPermission(ctx, input.WorkspaceID, input.ActorUserID); err != nil {
		return nil, err
	}

	members, err := s.workspaceRepo.ListMembers(ctx, input.WorkspaceID)
	if err != nil {
		return nil, err
	}

	items := make([]WorkspaceMemberResult, 0, len(members))
	for _, member := range members {
		items = append(items, workspaceMemberToResult(member))
	}

	return &WorkspaceMemberListResult{
		Items: items,
	}, nil
}

func (s *WorkspaceService) UpdateMemberRole(
	ctx context.Context,
	input UpdateWorkspaceMemberRoleInput,
) (*WorkspaceMemberResult, error) {
	newRole, err := parseWorkspaceRole(input.Role)
	if err != nil {
		return nil, err
	}

	actorMember, err := s.RequireManagePermission(ctx, input.WorkspaceID, input.ActorUserID)
	if err != nil {
		return nil, err
	}

	targetMember, err := s.RequireMember(ctx, input.WorkspaceID, input.TargetUserID)
	if err != nil {
		return nil, err
	}

	if err := canUpdateWorkspaceMemberRole(actorMember, targetMember, newRole); err != nil {
		return nil, err
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.workspaceRepo.UpdateMemberRole(
			ctx,
			tx,
			input.WorkspaceID,
			input.TargetUserID,
			newRole,
		)
	})
	if err != nil {
		return nil, err
	}

	updatedMember, err := s.workspaceRepo.FindMemberWithUser(
		ctx,
		input.WorkspaceID,
		input.TargetUserID,
	)
	if err != nil {
		return nil, err
	}

	return ptr(workspaceMemberToResult(*updatedMember)), nil
}

func (s *WorkspaceService) RemoveMember(
	ctx context.Context,
	input RemoveWorkspaceMemberInput,
) error {
	actorMember, err := s.RequireManagePermission(ctx, input.WorkspaceID, input.ActorUserID)
	if err != nil {
		return err
	}

	targetMember, err := s.RequireMember(ctx, input.WorkspaceID, input.TargetUserID)
	if err != nil {
		return err
	}

	if err := canRemoveWorkspaceMember(actorMember, targetMember); err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.workspaceRepo.DeleteMember(
			ctx,
			tx,
			input.WorkspaceID,
			input.TargetUserID,
		)
	})
}

func parseWorkspaceRole(role string) (model.WorkspaceRole, error) {
	normalized := strings.ToLower(strings.TrimSpace(role))

	switch model.WorkspaceRole(normalized) {
	case model.WorkspaceRoleOwner,
		model.WorkspaceRoleAdmin,
		model.WorkspaceRoleMember,
		model.WorkspaceRoleViewer:
		return model.WorkspaceRole(normalized), nil

	default:
		return "", workspace.NewWorkspaceError(
			workspace.ErrorCodeInvalidRole,
			"Workspace 角色无效",
			workspace.ErrInvalidRole,
		)
	}
}

func canUpdateWorkspaceMemberRole(
	actor *model.WorkspaceMember,
	target *model.WorkspaceMember,
	newRole model.WorkspaceRole,
) error {
	if actor.UserID == target.UserID {
		return workspace.NewWorkspaceError(
			workspace.ErrorCodePermissionDenied,
			"不能修改自己的角色",
			workspace.ErrPermissionDenied,
		)
	}

	if target.Role == model.WorkspaceRoleOwner {
		return workspace.NewWorkspaceError(
			workspace.ErrorCodeOwnerOperationNotAllowed,
			"不能修改 owner 角色",
			workspace.ErrOwnerOperationNotAllowed,
		)
	}

	if newRole == model.WorkspaceRoleOwner {
		return workspace.NewWorkspaceError(
			workspace.ErrorCodeOwnerOperationNotAllowed,
			"不能通过该接口设置 owner 角色",
			workspace.ErrOwnerOperationNotAllowed,
		)
	}

	if actor.Role == model.WorkspaceRoleOwner {
		return nil
	}

	if actor.Role == model.WorkspaceRoleAdmin {
		if target.Role == model.WorkspaceRoleAdmin {
			return workspace.NewWorkspaceError(
				workspace.ErrorCodePermissionDenied,
				"admin 不能修改其他 admin 的角色",
				workspace.ErrPermissionDenied,
			)
		}

		if newRole == model.WorkspaceRoleAdmin {
			return workspace.NewWorkspaceError(
				workspace.ErrorCodePermissionDenied,
				"admin 不能把成员提升为 admin",
				workspace.ErrPermissionDenied,
			)
		}

		return nil
	}

	return workspace.NewWorkspaceError(
		workspace.ErrorCodePermissionDenied,
		"当前用户没有修改成员角色的权限",
		workspace.ErrPermissionDenied,
	)
}

func canRemoveWorkspaceMember(
	actor *model.WorkspaceMember,
	target *model.WorkspaceMember,
) error {
	if actor.UserID == target.UserID {
		return workspace.NewWorkspaceError(
			workspace.ErrorCodePermissionDenied,
			"不能移除自己",
			workspace.ErrPermissionDenied,
		)
	}

	if target.Role == model.WorkspaceRoleOwner {
		return workspace.NewWorkspaceError(
			workspace.ErrorCodeOwnerOperationNotAllowed,
			"不能移除 owner",
			workspace.ErrOwnerOperationNotAllowed,
		)
	}

	if actor.Role == model.WorkspaceRoleOwner {
		return nil
	}

	if actor.Role == model.WorkspaceRoleAdmin {
		if target.Role == model.WorkspaceRoleAdmin {
			return workspace.NewWorkspaceError(
				workspace.ErrorCodePermissionDenied,
				"admin 不能移除其他 admin",
				workspace.ErrPermissionDenied,
			)
		}

		return nil
	}

	return workspace.NewWorkspaceError(
		workspace.ErrorCodePermissionDenied,
		"当前用户没有移除成员的权限",
		workspace.ErrPermissionDenied,
	)
}

func workspaceMemberToResult(member model.WorkspaceMember) WorkspaceMemberResult {
	result := WorkspaceMemberResult{
		UserID:   member.UserID,
		Role:     string(member.Role),
		JoinedAt: member.CreatedAt,
	}

	if member.User != nil {
		result.Email = member.User.Email
		result.DisplayName = member.User.DisplayName
	}

	return result
}

func ptr[T any](value T) *T {
	return &value
}
