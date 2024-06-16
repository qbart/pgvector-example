-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_documents_embedding ON document_chunks USING hnsw (embedding vector_cosine_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_documents_embedding;
-- +goose StatementEnd
