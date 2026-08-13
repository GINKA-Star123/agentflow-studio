package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type KnowledgeDocumentSourceType string

const (
	KnowledgeDocumentSourceUpload KnowledgeDocumentSourceType = "upload"
	KnowledgeDocumentSourceURL    KnowledgeDocumentSourceType = "url"
	KnowledgeDocumentSourceManual KnowledgeDocumentSourceType = "manual"
)

type KnowledgeDocumentStatus string

const (
	KnowledgeDocumentStatusPending    KnowledgeDocumentStatus = "pending"
	KnowledgeDocumentStatusProcessing KnowledgeDocumentStatus = "processing"
	KnowledgeDocumentStatusParsed     KnowledgeDocumentStatus = "parsed"
	KnowledgeDocumentStatusChunked    KnowledgeDocumentStatus = "chunked"
	KnowledgeDocumentStatusEmbedded   KnowledgeDocumentStatus = "embedded"
	KnowledgeDocumentStatusFailed     KnowledgeDocumentStatus = "failed"
)

type KnowledgeDocument struct {
	BaseModel

	WorkspaceID     uuid.UUID `gorm:"type:uuid;not null;index" json:"workspace_id"`
	KnowledgeBaseID uuid.UUID `gorm:"type:uuid;not null;index" json:"knowledge_base_id"`

	Filename    string                      `gorm:"type:varchar(512);not null" json:"filename"`
	SourceType  KnowledgeDocumentSourceType `gorm:"type:varchar(32);not null;default:'upload'" json:"source_type"`
	ContentType string                      `gorm:"type:varchar(255);not null;default:''" json:"content_type"`
	SizeBytes   int64                       `gorm:"not null;default:0" json:"size_bytes"`

	Status KnowledgeDocumentStatus `gorm:"type:varchar(32);not null;default:'pending'" json:"status"`

	StorageKey string `gorm:"type:varchar(1024);not null;default:''" json:"storage_key"`
	Checksum   string `gorm:"type:varchar(128);not null;default:''" json:"checksum"`

	Metadata datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	Error    datatypes.JSON `gorm:"type:jsonb" json:"error,omitempty"`

	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	UpdatedBy uuid.UUID `gorm:"type:uuid;not null" json:"updated_by"`

	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Workspace     *Workspace       `gorm:"foreignKey:WorkspaceID" json:"workspace,omitempty"`
	KnowledgeBase *KnowledgeBase   `gorm:"foreignKey:KnowledgeBaseID" json:"knowledge_base,omitempty"`
	Creator       *User            `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Updater       *User            `gorm:"foreignKey:UpdatedBy" json:"updater,omitempty"`
	Chunks        []KnowledgeChunk `gorm:"foreignKey:DocumentID" json:"chunks,omitempty"`
}

func (KnowledgeDocument) TableName() string {
	return "knowledge_documents"
}

func (s KnowledgeDocumentStatus) IsTerminal() bool {
	return s == KnowledgeDocumentStatusEmbedded ||
		s == KnowledgeDocumentStatusFailed
}

func (d KnowledgeDocument) IsReadyForSearch() bool {
	return d.Status == KnowledgeDocumentStatusEmbedded
}
