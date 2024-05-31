package documents

import "time"

type Document struct {
	ID        uint64 `bun:",autoincrement,identity"`
	Title     string
	Metadata  map[string]any `bun:"type:jsonb"`
	CreatedAt time.Time
}

type DocumentChunk struct {
	DocumentID uint64
	Chunk      uint64
	Chunks     uint64
	Content    string
	Embedding  []float32
	Metadata   map[string]any `bun:"type:jsonb"`
	Status     string
	UpdatedAt  time.Time
	CreatedAt  time.Time
}
