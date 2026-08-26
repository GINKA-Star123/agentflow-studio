package model

import "github.com/google/uuid"

type WorkspaceRole string

const (
	WorkspaceRoleOwner  WorkspaceRole = "owner"
	WorkspaceRoleAdmin  WorkspaceRole = "admin"
	WorkspaceRoleMember WorkspaceRole = "member"
	WorkspaceRoleViewer WorkspaceRole = "viewer"
)

type WorkspaceMember struct {
	BaseModel

	WorkspaceID uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:idx_workspace_member_unique,priority:1" json:"workspace_id"`
	UserID      uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:idx_workspace_member_unique,priority:2" json:"user_id"`
	Role        WorkspaceRole `gorm:"type:varchar(32);not null;default:'member';index" json:"role"`

	Workspace *Workspace `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (WorkspaceMember) TableName() string {
	return "workspace_members"
}


func (m WorkspaceMember) IsOwner() bool {
	return m.Role == WorkspaceRoleOwner
}

func (m WorkspaceMember) IsAdmin() bool {
	return m.Role == WorkspaceRoleAdmin || m.Role == WorkspaceRoleOwner
}

func (m WorkspaceMember) CanManageWorkspace() bool {
	return m.Role == WorkspaceRoleOwner || m.Role == WorkspaceRoleAdmin
}

func (m WorkspaceMember) CanViewWorkspace() bool {
	return m.Role == WorkspaceRoleOwner ||
		m.Role == WorkspaceRoleAdmin ||
		m.Role == WorkspaceRoleMember ||
		m.Role == WorkspaceRoleViewer
}
