DROP INDEX IF EXISTS idx_knowledge_chunks_vector_id_active;
DROP INDEX IF EXISTS idx_knowledge_chunks_vector_id;
DROP INDEX IF EXISTS idx_knowledge_chunks_workspace_kb;
DROP INDEX IF EXISTS idx_knowledge_chunks_document_id;
DROP INDEX IF EXISTS idx_knowledge_chunks_knowledge_base_id;
DROP INDEX IF EXISTS idx_knowledge_chunks_workspace_id;
DROP TABLE IF EXISTS knowledge_chunks;

DROP INDEX IF EXISTS idx_knowledge_documents_kb_checksum_active;
DROP INDEX IF EXISTS idx_knowledge_documents_workspace_kb_created_at;
DROP INDEX IF EXISTS idx_knowledge_documents_deleted_at;
DROP INDEX IF EXISTS idx_knowledge_documents_status;
DROP INDEX IF EXISTS idx_knowledge_documents_knowledge_base_id;
DROP INDEX IF EXISTS idx_knowledge_documents_workspace_id;
DROP TABLE IF EXISTS knowledge_documents;

DROP INDEX IF EXISTS idx_knowledge_bases_workspace_name_active;
DROP INDEX IF EXISTS idx_knowledge_bases_workspace_updated_at;
DROP INDEX IF EXISTS idx_knowledge_bases_deleted_at;
DROP INDEX IF EXISTS idx_knowledge_bases_workspace_id;
DROP TABLE IF EXISTS knowledge_bases;