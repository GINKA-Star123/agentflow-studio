CREATE TABLE IF NOT EXISTS knowledge_bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    workspace_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',

    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_knowledge_bases_workspace
        FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_knowledge_bases_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_knowledge_bases_updated_by
        FOREIGN KEY (updated_by)
        REFERENCES users(id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_knowledge_bases_workspace_id
ON knowledge_bases(workspace_id);

CREATE INDEX IF NOT EXISTS idx_knowledge_bases_deleted_at
ON knowledge_bases(deleted_at);

CREATE INDEX IF NOT EXISTS idx_knowledge_bases_workspace_updated_at
ON knowledge_bases(workspace_id, updated_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_knowledge_bases_workspace_name_active
ON knowledge_bases(workspace_id, LOWER(name))
WHERE deleted_at IS NULL;


CREATE TABLE IF NOT EXISTS knowledge_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    workspace_id UUID NOT NULL,
    knowledge_base_id UUID NOT NULL,

    filename VARCHAR(512) NOT NULL,
    source_type VARCHAR(32) NOT NULL DEFAULT 'upload',
    content_type VARCHAR(255) NOT NULL DEFAULT '',
    size_bytes BIGINT NOT NULL DEFAULT 0,

    status VARCHAR(32) NOT NULL DEFAULT 'pending',

    storage_key VARCHAR(1024) NOT NULL DEFAULT '',
    checksum VARCHAR(128) NOT NULL DEFAULT '',

    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    error JSONB,

    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT fk_knowledge_documents_workspace
        FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_knowledge_documents_knowledge_base
        FOREIGN KEY (knowledge_base_id)
        REFERENCES knowledge_bases(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_knowledge_documents_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_knowledge_documents_updated_by
        FOREIGN KEY (updated_by)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT knowledge_documents_source_type_check
        CHECK (source_type IN ('upload', 'url', 'manual')),

    CONSTRAINT knowledge_documents_status_check
        CHECK (status IN ('pending', 'processing', 'parsed', 'chunked', 'embedded', 'failed')),

    CONSTRAINT knowledge_documents_size_bytes_check
        CHECK (size_bytes >= 0)
);

CREATE INDEX IF NOT EXISTS idx_knowledge_documents_workspace_id
ON knowledge_documents(workspace_id);

CREATE INDEX IF NOT EXISTS idx_knowledge_documents_knowledge_base_id
ON knowledge_documents(knowledge_base_id);

CREATE INDEX IF NOT EXISTS idx_knowledge_documents_status
ON knowledge_documents(status);

CREATE INDEX IF NOT EXISTS idx_knowledge_documents_deleted_at
ON knowledge_documents(deleted_at);

CREATE INDEX IF NOT EXISTS idx_knowledge_documents_workspace_kb_created_at
ON knowledge_documents(workspace_id, knowledge_base_id, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_knowledge_documents_kb_checksum_active
ON knowledge_documents(knowledge_base_id, checksum)
WHERE deleted_at IS NULL AND checksum <> '';


CREATE TABLE IF NOT EXISTS knowledge_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    workspace_id UUID NOT NULL,
    knowledge_base_id UUID NOT NULL,
    document_id UUID NOT NULL,

    chunk_index INTEGER NOT NULL,
    content TEXT NOT NULL,
    token_count INTEGER NOT NULL DEFAULT 0,

    embedding_model VARCHAR(255) NOT NULL DEFAULT '',
    vector_id VARCHAR(255) NOT NULL DEFAULT '',

    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_knowledge_chunks_workspace
        FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_knowledge_chunks_knowledge_base
        FOREIGN KEY (knowledge_base_id)
        REFERENCES knowledge_bases(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_knowledge_chunks_document
        FOREIGN KEY (document_id)
        REFERENCES knowledge_documents(id)
        ON DELETE CASCADE,

    CONSTRAINT knowledge_chunks_chunk_index_check
        CHECK (chunk_index >= 0),

    CONSTRAINT knowledge_chunks_token_count_check
        CHECK (token_count >= 0),

    CONSTRAINT knowledge_chunks_document_index_unique
        UNIQUE (document_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_workspace_id
ON knowledge_chunks(workspace_id);

CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_knowledge_base_id
ON knowledge_chunks(knowledge_base_id);

CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_document_id
ON knowledge_chunks(document_id);

CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_workspace_kb
ON knowledge_chunks(workspace_id, knowledge_base_id);

CREATE INDEX IF NOT EXISTS idx_knowledge_chunks_vector_id
ON knowledge_chunks(vector_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_knowledge_chunks_vector_id_active
ON knowledge_chunks(vector_id)
WHERE vector_id <> '';