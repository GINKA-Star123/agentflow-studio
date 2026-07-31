package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type NodeExecutionStatus string

const (
	NodeExecutionStatusPending   NodeExecutionStatus = "pending"
	NodeExecutionStatusRunning   NodeExecutionStatus = "running"
	NodeExecutionStatusSucceeded NodeExecutionStatus = "succeeded"
	NodeExecutionStatusFailed    NodeExecutionStatus = "failed"
	NodeExecutionStatusSkipped   NodeExecutionStatus = "skipped"
)

type NodeExecution struct {
	BaseModel

	WorkspaceID uuid.UUID `gorm:"type:uuid;not null;index" json:"workspace_id"`
	RunID       uuid.UUID `gorm:"type:uuid;not null;index" json:"run_id"`

	NodeID   string `gorm:"type:varchar(128);not null;index" json:"node_id"`
	NodeType string `gorm:"type:varchar(64);not null;index" json:"node_type"`
	Sequence int    `gorm:"not null;default:0" json:"sequence"`

	Status NodeExecutionStatus `gorm:"type:varchar(32);not null;default:'pending';index" json:"status"`

	Input      datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"input"`
	Output     datatypes.JSON `gorm:"type:jsonb" json:"output,omitempty"`
	Error      datatypes.JSON `gorm:"type:jsonb" json:"error,omitempty"`
	TokenUsage datatypes.JSON `gorm:"type:jsonb" json:"token_usage,omitempty"`

	LatencyMS int64 `gorm:"not null;default:0" json:"latency_ms"`

	StartedAt  *time.Time `gorm:"type:timestamptz" json:"started_at,omitempty"`
	FinishedAt *time.Time `gorm:"type:timestamptz" json:"finished_at,omitempty"`

	Run *WorkflowRun `gorm:"foreignKey:RunID" json:"run,omitempty"`
}

func (NodeExecution) TableName() string {
	return "node_execution"
}

func (s NodeExecutionStatus) IsTerminal() bool {
	return s == NodeExecutionStatusSucceeded ||
		s == NodeExecutionStatusFailed ||
		s == NodeExecutionStatusSkipped
}

func (e NodeExecution) IsTerminal() bool {
	return e.Status.IsTerminal()
}

func (e *NodeExecution) MarkRunning(now time.Time, input datatypes.JSON) {
	e.Status = NodeExecutionStatusRunning
	e.Input = input
	e.StartedAt = &now
}

func (e *NodeExecution) MarkSucceeded(now time.Time, output datatypes.JSON, latencyMS int64) {
	e.Status = NodeExecutionStatusSucceeded
	e.Output = output
	e.FinishedAt = &now
	e.LatencyMS = latencyMS
}

func (e *NodeExecution) MarkFailed(now time.Time, errorPayload datatypes.JSON, latencyMS int64) {
	e.Status = NodeExecutionStatusFailed
	e.Error = errorPayload
	e.FinishedAt = &now
	e.LatencyMS = latencyMS
}

func (e *NodeExecution) MarkSkipped(now time.Time, reason datatypes.JSON) {
	e.Status = NodeExecutionStatusSkipped
	e.Error = reason
	e.FinishedAt = &now
}
