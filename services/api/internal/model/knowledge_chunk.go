package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type KnowledgeChunk struct {
	BaseModel

	WorkspaceID     uuid.UUID `gorm:"type:uuid;not null;index" json:"workspace_id"`
	KnowledgeBaseID uuid.UUID `gorm:"type:uuid;not null;index" json:"knowledge_base_id"`
	DocumentID      uuid.UUID `gorm:"type:uuid;not null;index" json:"document_id"`

	ChunkIndex int    `gorm:"not null" json:"chunk_index"`
	Content    string `gorm:"type:text;not null" json:"content"`
	TokenCount int    `gorm:"not null;default:0" json:"token_count"`

	EmbeddingModel string `gorm:"type:varchar(255);not null;default:''" json:"embedding_model"`
	VectorID       string `gorm:"type:varchar(255);not null;default:''" json:"vector_id"`

	Metadata datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`

	Workspace     *Workspace         `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
	KnowledgeBase *KnowledgeBase     `gorm:"foreignKey:KnowledgeBaseID" json:"knowledge_base,omitempty"`
	Document      *KnowledgeDocument `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
}

func (KnowledgeChunk) TableName() string {
	return "knowledge_chunks"
}
