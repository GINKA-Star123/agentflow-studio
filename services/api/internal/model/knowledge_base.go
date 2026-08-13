package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KnowledgeBase struct {
	BaseModel

	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text;not null;default:''" json:"description"`

	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	UpdatedBy uuid.UUID `gorm:"type:uuid;not null" json:"updated_by"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Workspace *Workspace          `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
	Creator   *User               `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater   *User               `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	Documents []KnowledgeDocument `gorm:"foreignKey:KnowledgeBaseID" json:"documents,omitempty"`
}

func (KnowledgeBase) TableName() string {
	return "knowledge_bases"
}
