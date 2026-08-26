package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type WorkflowRunStatus string

const (
	WorkflowRunStatusPending   WorkflowRunStatus = "pending"
	WorkflowRunStatusRunning   WorkflowRunStatus = "running"
	WorkflowRunStatusSucceeded WorkflowRunStatus = "succeeded"
	WorkflowRunStatusFailed    WorkflowRunStatus = "failed"
	WorkflowRunStatusCanceled  WorkflowRunStatus = "canceled"
)

type WorkflowRun struct {
	BaseModel

	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index" json:"workspace_id"`
	WorkflowID  uuid.UUID `gorm:"type:uuid;not null;index" json:"workflow_id"`
	TriggeredBy uuid.UUID `gorm:"type:uuid;not null;index" json:"triggered_by"`

	Status WorkflowRunStatus `gorm:"type:varchar(32);not null;default:'pending'" json:"status"`

	SchemaSnapshot datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"schema_snapshot"`
	Input          datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"input"`
	Output         datatypes.JSON `gorm:"type:jsonb" json:"output,omitempty"`
	Error          datatypes.JSON `gorm:"type:jsonb" json:"error,omitempty"`

	StartedAt  *time.Time `gorm:"type:timestamptz" json:"started_at,omitempty"`
	FinishedAt *time.Time `gorm:"type:timestamptz" json:"finished_at,omitempty"`

	Workspace *Workspace `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
	Triggerer *User      `gorm:"foreignKey:TriggeredBy" json:"triggered_by_user,omitempty"`
}

func (WorkflowRun) TableName() string {
	return "workflow_runs"
}

func (s WorkflowRunStatus) IsTerminal() bool {
	return s == WorkflowRunStatusSucceeded ||
		s == WorkflowRunStatusFailed ||
		s == WorkflowRunStatusCanceled
}

func (r WorkflowRun) IsTerminal() bool {
	return r.Status.IsTerminal()
}

func (r *WorkflowRun) MarkRunning(now time.Time) {
	r.Status = WorkflowRunStatusRunning
	r.StartedAt = &now
}

func (r *WorkflowRun) MarkSucceeded(now time.Time, output datatypes.JSON) {
	r.Status = WorkflowRunStatusSucceeded
	r.Output = output
	r.FinishedAt = &now
}

func (r *WorkflowRun) MarkFailed(now time.Time, errorPayload datatypes.JSON) {
	r.Status = WorkflowRunStatusFailed
	r.Error = errorPayload
	r.FinishedAt = &now
}

func (r *WorkflowRun) MarkCanceled(now time.Time, errorPayload datatypes.JSON) {
	r.Status = WorkflowRunStatusCanceled
	r.Error = errorPayload
	r.FinishedAt = &now
}
