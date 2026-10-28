package repository

import (
	"context"

	"agentflow-studio/services/api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type KnowledgeRepository struct {
	db *gorm.DB
}

func NewKnowledgeRepository(db *gorm.DB) *KnowledgeRepository {
	return &KnowledgeRepository{
		db: db,
	}
}

func (r *KnowledgeRepository) ListKnowledgeBases(
	ctx context.Context,
	workspaceID uuid.UUID,
) ([]model.KnowledgeBase, error) {
	var items []model.KnowledgeBase

	err := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("updated_at DESC").
		Find(&items).
		Error

	return items, err
}

func (r *KnowledgeRepository) FindKnowledgeBaseByID(
	ctx context.Context,
	workspaceID uuid.UUID,
	knowledgeBaseID uuid.UUID,
) (*model.KnowledgeBase, error) {
	var item model.KnowledgeBase

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND id = ?", workspaceID, knowledgeBaseID).
		First(&item).
		Error

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *KnowledgeRepository) CreateKnowledgeBase(
	ctx context.Context,
	tx *gorm.DB,
	item *model.KnowledgeBase,
) error {
	return r.useDB(tx).WithContext(ctx).Create(item).Error
}

func (r *KnowledgeRepository) UpdateKnowledgeBase(
	ctx context.Context,
	tx *gorm.DB,
	item *model.KnowledgeBase,
) error {
	result := r.useDB(tx).WithContext(ctx).
		Model(&model.KnowledgeBase{}).
		Where("workspace_id = ? AND id = ?", item.WorkspaceID, item.ID).
		Updates(map[string]any{
			"name":        item.Name,
			"description": item.Description,
			"updated_by":  item.UpdatedBy,
			"updated_at":  item.UpdatedAt,
		})

	return requireRowsAffected(result)
}

func (r *KnowledgeRepository) DeleteKnowledgeBase(
	ctx context.Context,
	tx *gorm.DB,
	workspaceID uuid.UUID,
	knowledgeBaseID uuid.UUID,
) error {
	result := r.useDB(tx).WithContext(ctx).
		Where("workspace_id = ? AND id = ?", workspaceID, knowledgeBaseID).
		Delete(&model.KnowledgeBase{})

	return requireRowsAffected(result)
}

func (r *KnowledgeRepository) CreateDocument(
	ctx context.Context,
	tx *gorm.DB,
	document *model.KnowledgeDocument,
) error {
	return r.useDB(tx).WithContext(ctx).Create(document).Error
}

func (r *KnowledgeRepository) ListDocuments(
	ctx context.Context,
	workspaceID uuid.UUID,
	knowledgeBaseID uuid.UUID,
	limit int,
) ([]model.KnowledgeDocument, error) {
	var items []model.KnowledgeDocument

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND knowledge_base_id = ?", workspaceID, knowledgeBaseID).
		Order("created_at DESC").
		Limit(normalizeLimit(limit, 50, 200)).
		Find(&items).
		Error

	return items, err
}

func (r *KnowledgeRepository) FindDocumentByID(
	ctx context.Context,
	workspaceID uuid.UUID,
	knowledgeBaseID uuid.UUID,
	documentID uuid.UUID,
) (*model.KnowledgeDocument, error) {
	var item model.KnowledgeDocument

	err := r.db.WithContext(ctx).
		Where(
			"workspace_id = ? AND knowledge_base_id = ? AND id = ?",
			workspaceID,
			knowledgeBaseID,
			documentID,
		).
		First(&item).
		Error

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *KnowledgeRepository) UpdateDocument(
	ctx context.Context,
	tx *gorm.DB,
	document *model.KnowledgeDocument,
) error {
	result := r.useDB(tx).WithContext(ctx).
		Model(&model.KnowledgeDocument{}).
		Where(
			"workspace_id = ? AND knowledge_base_id = ? AND id = ?",
			document.WorkspaceID,
			document.KnowledgeBaseID,
			document.ID,
		).
		Updates(map[string]any{
			"filename":     document.Filename,
			"source_type":  document.SourceType,
			"content_type": document.ContentType,
			"size_bytes":   document.SizeBytes,
			"status":       document.Status,
			"storage_key":  document.StorageKey,
			"checksum":     document.Checksum,
			"metadata":     document.Metadata,
			"error":        document.Error,
			"updated_by":   document.UpdatedBy,
			"updated_at":   document.UpdatedAt,
		})

	return requireRowsAffected(result)
}

func (r *KnowledgeRepository) UpdateDocumentStatus(
	ctx context.Context,
	tx *gorm.DB,
	document *model.KnowledgeDocument,
) error {
	result := r.useDB(tx).WithContext(ctx).
		Model(&model.KnowledgeDocument{}).
		Where("workspace_id = ? AND id = ?", document.WorkspaceID, document.ID).
		Updates(map[string]any{
			"status":     document.Status,
			"error":      document.Error,
			"updated_by": document.UpdatedBy,
			"updated_at": document.UpdatedAt,
		})

	return requireRowsAffected(result)
}

func (r *KnowledgeRepository) DeleteDocument(
	ctx context.Context,
	tx *gorm.DB,
	workspaceID uuid.UUID,
	knowledgeBaseID uuid.UUID,
	documentID uuid.UUID,
) error {
	result := r.useDB(tx).WithContext(ctx).
		Where(
			"workspace_id = ? AND knowledge_base_id = ? AND id = ?",
			workspaceID,
			knowledgeBaseID,
			documentID,
		).
		Delete(&model.KnowledgeDocument{})

	return requireRowsAffected(result)
}

func (r *KnowledgeRepository) CreateChunks(
	ctx context.Context,
	tx *gorm.DB,
	chunks []model.KnowledgeChunk,
) error {
	if len(chunks) == 0 {
		return nil
	}

	return r.useDB(tx).WithContext(ctx).
		CreateInBatches(chunks, 100).
		Error
}

func (r *KnowledgeRepository) ListChunksByDocument(
	ctx context.Context,
	workspaceID uuid.UUID,
	documentID uuid.UUID,
) ([]model.KnowledgeChunk, error) {
	var items []model.KnowledgeChunk

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND document_id = ?", workspaceID, documentID).
		Order("chunk_index ASC").
		Find(&items).
		Error

	return items, err
}

func (r *KnowledgeRepository) ListChunksByKnowledgeBase(
	ctx context.Context,
	workspaceID uuid.UUID,
	knowledgeBaseID uuid.UUID,
	limit int,
) ([]model.KnowledgeChunk, error) {
	var items []model.KnowledgeChunk

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND knowledge_base_id = ?", workspaceID, knowledgeBaseID).
		Order("created_at DESC").
		Limit(normalizeLimit(limit, 100, 500)).
		Find(&items).
		Error

	return items, err
}

func (r *KnowledgeRepository) FindChunkByID(
	ctx context.Context,
	workspaceID uuid.UUID,
	chunkID uuid.UUID,
) (*model.KnowledgeChunk, error) {
	var item model.KnowledgeChunk

	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND id = ?", workspaceID, chunkID).
		First(&item).
		Error

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *KnowledgeRepository) DeleteChunksByDocument(
	ctx context.Context,
	tx *gorm.DB,
	workspaceID uuid.UUID,
	documentID uuid.UUID,
) error {
	return r.useDB(tx).WithContext(ctx).
		Where("workspace_id = ? AND document_id = ?", workspaceID, documentID).
		Delete(&model.KnowledgeChunk{}).
		Error
}

func (r *KnowledgeRepository) useDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}

	return r.db
}

func requireRowsAffected(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func normalizeLimit(limit int, defaultLimit int, maxLimit int) int {
	if limit <= 0 {
		return defaultLimit
	}

	if limit > maxLimit {
		return maxLimit
	}

	return limit
}
