package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Workflow 保存一个 Workspace 下的当前 Workflow 定义。
// SchemaJSON 保留完整的画布 Schema，运行时会从该字段读取定义。
type Workflow struct {
	BaseModel

	WorkspaceID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"workspace_id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	SchemaVersion string         `gorm:"type:varchar(32);not null;default:'1.0'" json:"schema_version"`
	SchemaJSON    datatypes.JSON `gorm:"type:jsonb;not null" json:"-"`
	CreatedBy     uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	UpdatedBy     uuid.UUID      `gorm:"type:uuid;not null" json:"updated_by"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Workflow) TableName() string {
	return "workflows"
}
