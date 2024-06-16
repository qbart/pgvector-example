package documents

import (
	"context"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/pgvector/pgvector-go"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/uptrace/bun"
)

const DocumentUploadLimit = 10 * 1024 * 1024 // 10 MB

type Service interface {
	SearchDocument(ctx context.Context, query string, score float32) ([]SearchResponse, error)
	IndexDocument(ctx context.Context, filename string, filesize int64, data multipart.File) error
	Ask(ctx context.Context, input *AskInput) (*AskOutput, error)
}

type SearchResponse struct {
	Title           string
	DocumentID      int64
	Chunk           int64
	Content         string
	Page            int64
	Score           float32
	AcceptableScore bool
}

type AskOutput struct {
	Answer string
}

type AskInput struct {
	Question string
	Context  []string
}

type service struct {
	db     *bun.DB
	openAI *OpenAI
}

func NewService(db *bun.DB, openAI *OpenAI) *service {
	return &service{
		db:     db,
		openAI: openAI,
	}
}

func (s *service) SearchDocument(ctx context.Context, query string, score float32) ([]SearchResponse, error) {
	embedded, err := s.openAI.CreateEmbedding(ctx, query)
	if err != nil {
		slog.Error("Error creating embedding", "message", err.Error())
		return nil, err
	}

	var chunks []SearchResponse
	err = s.db.NewSelect().
		TableExpr("document_chunks AS dc").
		ColumnExpr("d.title, dc.document_id, dc.chunk, dc.content, (dc.metadata->>'page')::bigint as page").
		ColumnExpr("1 - (dc.embedding <=> ?) AS score", pgvector.NewVector(embedded)).
		Join("JOIN documents d ON d.id = dc.document_id").
		Limit(10).
		OrderExpr("score DESC").
		Scan(ctx, &chunks)
	if err != nil {
		slog.Error("Error searching documents", "message", err.Error())
		return nil, err
	}

	for i := range chunks {
		chunks[i].AcceptableScore = chunks[i].Score > score
	}
	return chunks, nil
}

func (s *service) IndexDocument(ctx context.Context, filename string, filesize int64, data multipart.File) error {
	doc := &Document{
		Title: filename,
		Metadata: map[string]any{
			"filename": filename,
			"type":     mime.TypeByExtension(filepath.Ext(filename)),
			"size":     filesize,
		},
		CreatedAt: time.Now().UTC(),
	}
	_, err := s.db.NewInsert().Model(doc).Exec(ctx)
	if err != nil {
		return err
	}

	go s.processDocument(ctx, doc.ID, data, filesize)

	return nil
}

func (s *service) processDocument(ctx context.Context, id int64, file io.Reader, size int64) {
	splitter := textsplitter.NewTokenSplitter(
		textsplitter.WithModelName("text-embedding-3-small"),
		textsplitter.WithChunkSize(500),
		textsplitter.WithChunkOverlap(75), // 15% of chunk size
	)

	pdf := documentloaders.NewText(file)
	// pdf := documentloaders.NewPDF(file, size)
	docs, err := pdf.LoadAndSplit(ctx, splitter)
	if err != nil {
		slog.Error("Error loading and splitting document", "message", err.Error())
		return
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Error beginning transaction", "message", err.Error())
		return
	}
	chunks := len(docs)
	slog.Info("Split document", "document_id", id, "chunks", chunks)
	// for _, doc := range docs {
	// 	fmt.Println(doc.PageContent)
	// }
	// return
	for i, doc := range docs {
		embedded, err := s.openAI.CreateEmbedding(ctx, doc.PageContent)
		if err != nil {
			slog.Error("Error creating embedding", "message", err.Error())
			tx.Rollback()
			return
		}
		slog.Info("Processing document chunk", "document_id", id, "chunk", i+1)

		chunk := &DocumentChunk{
			DocumentID: id,
			Chunk:      int64(i + 1),
			Chunks:     int64(chunks),
			Content:    doc.PageContent,
			Metadata:   doc.Metadata,
			Embedding:  pgvector.NewVector(embedded),
			Status:     "success",
		}

		_, err = tx.NewInsert().Model(chunk).Exec(ctx)
		if err != nil {
			slog.Error("Error inserting chunk", "message", err.Error())
			tx.Rollback()
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		slog.Error("Error committing transaction", "message", err.Error())
		return
	}
	slog.Info("Processing document finished", "document_id", id)
}

func (s *service) Ask(ctx context.Context, input *AskInput) (*AskOutput, error) {
	enhancedContext := ""
	for _, s := range input.Context {
		enhancedContext += s + "\n\n"
	}
	fullMsg := "Kontekst:\n\n" + enhancedContext + "-------\n\nPytanie: " + input.Question + "\n\nJeśli kontekst nie jest wystarczający, napisz, że nie możesz znaleźć odpowiedzi w dokumentach."
	msg := llms.MessageContent{
		Role: llms.ChatMessageTypeGeneric,
		Parts: []llms.ContentPart{
			llms.TextPart(fullMsg),
		},
	}
	resp, err := s.openAI.llm.GenerateContent(ctx, []llms.MessageContent{msg})
	if err != nil {
		slog.Error("Error generating content", "message", err.Error())
		return nil, err
	}
	out := &AskOutput{
		Answer: resp.Choices[0].Content,
	}
	return out, nil
}
