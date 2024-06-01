package documents

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"

	"github.com/pgvector/pgvector-go"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/uptrace/bun"
)

const DocumentUploadLimit = 10 * 1024 * 1024 // 10 MB

type HttpHandler struct {
	DB *bun.DB
	*OpenAI
}

func (h *HttpHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(DocumentUploadLimit)
	form := r.MultipartForm
	file := form.File["file"]
	if len(file) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no file uploaded"))
	}

	for _, fileHeader := range file {
		src, err := fileHeader.Open()
		if err != nil {
			slog.Error("Error opening file", "message", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer src.Close()

		id, err := h.saveDocument(r.Context(), fileHeader.Filename, fileHeader.Size, src)
		if err != nil {
			slog.Error("Error saving document", "message", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"id": id})
	}
}

func (h *HttpHandler) saveDocument(ctx context.Context, name string, size int64, file multipart.File) (int64, error) {
	doc := &Document{
		Title: name,
		Metadata: map[string]any{
			"filename": name,
			"type":     mime.TypeByExtension(filepath.Ext(name)),
			"size":     size,
		},
		CreatedAt: time.Now(),
	}
	_, err := h.DB.NewInsert().Model(doc).Exec(ctx)
	if err != nil {
		return 0, err
	}

	go h.processDocument(doc.ID, file, size)

	return doc.ID, nil
}

func (h *HttpHandler) processDocument(id int64, file io.ReaderAt, size int64) {
	splitter := textsplitter.NewTokenSplitter(
		textsplitter.WithModelName("text-embedding-3-small"),
		textsplitter.WithChunkSize(500),
		textsplitter.WithChunkOverlap(75), // 15% of chunk size
		textsplitter.WithKeepSeparator(true),
		textsplitter.WithCodeBlocks(true),
	)

	pdf := documentloaders.NewPDF(file, size)
	docs, err := pdf.LoadAndSplit(context.TODO(), splitter)
	if err != nil {
		slog.Error("Error loading and splitting document", "message", err.Error())
		return
	}
	tx, err := h.DB.BeginTx(context.TODO(), nil)
	if err != nil {
		slog.Error("Error beginning transaction", "message", err.Error())
		return
	}
	chunks := len(docs)
	slog.Info("Split document", "document_id", id, "chunks", chunks)
	for i, doc := range docs {
		embedded, err := h.OpenAI.CreateEmbedding(context.TODO(), doc.PageContent)
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

		_, err = tx.NewInsert().Model(chunk).Exec(context.TODO())
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
