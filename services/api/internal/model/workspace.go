package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Workspace struct {
	BaseModel

	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	OwnerID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"owner_id"`
	Owner     *User          `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Workspace) TableName() string {
	return "workspaces"
}

func (w Workspace) IsDeleted() bool {
	return w.DeletedAt.Valid && !w.DeletedAt.Time.IsZero()
}

func (w Workspace) DeletedAtTime() *time.Time {
	if !w.DeletedAt.Valid {
		return nil
	}
	return &w.DeletedAt.Time
}
