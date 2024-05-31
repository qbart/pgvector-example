-- +goose Up
-- +goose StatementBegin
CREATE TABLE document_chunks (
  document_id BIGINT NOT NULL,
  chunk BIGINT NOT NULL,
  chunks BIGINT NOT NULL,
  content TEXT NOT NULL,
  embedding VECTOR(1536) NOT NULL,
  metadata JSONB NOT NULL DEFAULT '{}',
  status VARCHAR(255) NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (document_id, chunk)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE document_chunks;
-- +goose StatementEnd
