package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"agentflow-studio/services/api/internal/auth"
	"agentflow-studio/services/api/internal/model"
	"agentflow-studio/services/api/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	db            *gorm.DB
	userRepo      *repository.UserRepository
	workspaceRepo *repository.WorkspaceRepository
	jwtManager    *auth.JWTManager
}

func NewAuthService(
	db *gorm.DB,
	userRepo *repository.UserRepository,
	workspaceRepo *repository.WorkspaceRepository,
	jwtManager *auth.JWTManager,
) *AuthService {
	return &AuthService{
		db:            db,
		userRepo:      userRepo,
		workspaceRepo: workspaceRepo,
		jwtManager:    jwtManager,
	}
}

type RegisterInput struct {
	Email         string
	Password      string
	DisplayName   string
	WorkspaceName string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthResult struct {
	User             UserDTO        `json:"user"`
	AccessToken      string         `json:"access_token"`
	TokenType        string         `json:"token_type"`
	ExpiresAt        time.Time      `json:"expires_at"`
	CurrentWorkspace *WorkspaceDTO  `json:"current_workspace,omitempty"`
	Workspaces       []WorkspaceDTO `json:"workspaces"`
}

type MeResult struct {
	User             UserDTO        `json:"user"`
	CurrentWorkspace *WorkspaceDTO  `json:"current_workspace,omitempty"`
	Workspaces       []WorkspaceDTO `json:"workspaces"`
}

type UserDTO struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
}

type WorkspaceDTO struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	OwnerID uuid.UUID `json:"owner_id"`
	Role    string    `json:"role,omitempty"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	email := normalizeEmail(input.Email)
	displayName := strings.TrimSpace(input.DisplayName)
	workspaceName := strings.TrimSpace(input.WorkspaceName)

	if email == "" {
		return nil, auth.NewAuthError(
			auth.ErrorCodeInvalidInput,
			"邮箱不能为空",
			auth.ErrInvalidInput,
		)
	}

	if displayName == "" {
		displayName = defaultDisplayName(email)
	}

	if workspaceName == "" {
		workspaceName = fmt.Sprintf("%s 的工作区", displayName)
	}

	if err := auth.ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	exists, err := s.userRepo.EmailExists(ctx, email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, auth.NewAuthError(
			auth.ErrorCodeEmailAlreadyExists,
			"邮箱已注册",
			auth.ErrEmailAlreadyExists,
		)
	}

	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	var createdUser *model.User
	var createdWorkspace *model.Workspace

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := &model.User{
			Email:        email,
			PasswordHash: passwordHash,
			DisplayName:  displayName,
			Status:       model.UserStatusActive,
		}

		if err := s.userRepo.Create(ctx, tx, user); err != nil {
			return err
		}

		workspace := &model.Workspace{
			Name:    workspaceName,
			OwnerID: user.ID,
		}

		if err := s.workspaceRepo.Create(ctx, tx, workspace); err != nil {
			return err
		}

		member := &model.WorkspaceMember{
			WorkspaceID: workspace.ID,
			UserID:      user.ID,
			Role:        model.WorkspaceRoleOwner,
		}

		if err := s.workspaceRepo.CreateMember(ctx, tx, member); err != nil {
			return err
		}

		createdUser = user
		createdWorkspace = workspace

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.buildAuthResult(ctx, createdUser, createdWorkspace)
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	email := normalizeEmail(input.Email)

	if email == "" || strings.TrimSpace(input.Password) == "" {
		return nil, auth.NewAuthError(
			auth.ErrorCodeInvalidCredentials,
			"邮箱或密码错误",
			auth.ErrInvalidCredentials,
		)
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, auth.NewAuthError(
				auth.ErrorCodeInvalidCredentials,
				"邮箱或密码错误",
				auth.ErrInvalidCredentials,
			)
		}

		return nil, err
	}

	if user.Status != model.UserStatusActive {
		return nil, auth.NewAuthError(
			auth.ErrorCodeUserDisabled,
			"用户已被禁用",
			auth.ErrUserDisabled,
		)
	}

	if err := auth.ComparePassword(user.PasswordHash, input.Password); err != nil {
		return nil, err
	}

	workspaces, err := s.workspaceRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	var currentWorkspace *model.Workspace
	if len(workspaces) > 0 {
		currentWorkspace = &workspaces[0]
	}

	return s.buildAuthResult(ctx, user, currentWorkspace)
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*MeResult, error) {
	if userID == uuid.Nil {
		return nil, auth.NewAuthError(
			auth.ErrorCodeInvalidToken,
			"当前用户 ID 无效",
			auth.ErrInvalidToken,
		)
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, auth.NewAuthError(
				auth.ErrorCodeInvalidToken,
				"当前用户不存在",
				auth.ErrInvalidToken,
			)
		}

		return nil, err
	}

	if user.Status != model.UserStatusActive {
		return nil, auth.NewAuthError(
			auth.ErrorCodeUserDisabled,
			"用户已被禁用",
			auth.ErrUserDisabled,
		)
	}

	workspaces, err := s.workspaceRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	var currentWorkspace *model.Workspace
	if len(workspaces) > 0 {
		currentWorkspace = &workspaces[0]
	}

	workspaceDTOs, currentWorkspaceDTO, err := s.buildWorkspaceDTOs(ctx, user.ID, workspaces, currentWorkspace)
	if err != nil {
		return nil, err
	}

	return &MeResult{
		User: UserDTO{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Status:      string(user.Status),
		},
		CurrentWorkspace: currentWorkspaceDTO,
		Workspaces:       workspaceDTOs,
	}, nil
}

func (s *AuthService) buildAuthResult(
	ctx context.Context,
	user *model.User,
	currentWorkspace *model.Workspace,
) (*AuthResult, error) {
	accessToken, expiresAt, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	workspaces, err := s.workspaceRepo.ListByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	workspaceDTOs, currentWorkspaceDTO, err := s.buildWorkspaceDTOs(ctx, user.ID, workspaces, currentWorkspace)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User: UserDTO{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			Status:      string(user.Status),
		},
		AccessToken:      accessToken,
		TokenType:        "Bearer",
		ExpiresAt:        expiresAt,
		CurrentWorkspace: currentWorkspaceDTO,
		Workspaces:       workspaceDTOs,
	}, nil
}

func (s *AuthService) buildWorkspaceDTOs(
	ctx context.Context,
	userID uuid.UUID,
	workspaces []model.Workspace,
	currentWorkspace *model.Workspace,
) ([]WorkspaceDTO, *WorkspaceDTO, error) {
	workspaceDTOs := make([]WorkspaceDTO, 0, len(workspaces))

	for _, workspace := range workspaces {
		role, err := s.getWorkspaceRole(ctx, workspace.ID, userID)
		if err != nil {
			return nil, nil, err
		}

		workspaceDTOs = append(workspaceDTOs, WorkspaceDTO{
			ID:      workspace.ID,
			Name:    workspace.Name,
			OwnerID: workspace.OwnerID,
			Role:    role,
		})
	}

	var currentWorkspaceDTO *WorkspaceDTO
	if currentWorkspace != nil {
		role, err := s.getWorkspaceRole(ctx, currentWorkspace.ID, userID)
		if err != nil {
			return nil, nil, err
		}

		currentWorkspaceDTO = &WorkspaceDTO{
			ID:      currentWorkspace.ID,
			Name:    currentWorkspace.Name,
			OwnerID: currentWorkspace.OwnerID,
			Role:    role,
		}
	}

	return workspaceDTOs, currentWorkspaceDTO, nil
}

func (s *AuthService) getWorkspaceRole(
	ctx context.Context,
	workspaceID uuid.UUID,
	userID uuid.UUID,
) (string, error) {
	member, err := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
	if err != nil {
		return "", err
	}

	return string(member.Role), nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func defaultDisplayName(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "新用户"
	}

	return strings.TrimSpace(parts[0])
}
