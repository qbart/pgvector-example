package documents

import (
	"time"

	"github.com/pgvector/pgvector-go"
)

type Document struct {
	ID        int64 `bun:",autoincrement,identity"`
	Title     string
	Metadata  map[string]any `bun:"type:jsonb"`
	CreatedAt time.Time
}

type DocumentChunk struct {
	DocumentID int64
	Chunk      int64
	Chunks     int64
	Content    string
	Embedding  pgvector.Vector `bun:"type:vector(1536)"`
	Metadata   map[string]any  `bun:"type:jsonb"`
	Status     string
	UpdatedAt  time.Time
	CreatedAt  time.Time
}
