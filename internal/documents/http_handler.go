package documents

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/uptrace/bun"
)

const DocumentUploadLimit = 10 * 1024 * 1024 // 10 MB

type HttpHandler struct {
	DB *bun.DB
}

func (h *HttpHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(DocumentUploadLimit)
	form := r.MultipartForm
	file := form.File["file"]

	for _, fileHeader := range file {
		src, err := fileHeader.Open()
		if err != nil {
			slog.Error("Error opening file", "message", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer src.Close()

		id, err := h.saveDocument(r.Context(), fileHeader.Filename, src)
		if err != nil {
			slog.Error("Error saving document", "message", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"id": id})
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("no file uploaded"))
}

func (h *HttpHandler) saveDocument(ctx context.Context, name string, data io.Reader) (uint64, error) {
	doc := &Document{
		Title: name,
		Metadata: map[string]any{
			"filename": name,
			"type":     mime.TypeByExtension(filepath.Ext(name)),
		},
		CreatedAt: time.Now(),
	}
	_, err := h.DB.NewInsert().Model(doc).Exec(ctx)
	if err != nil {
		return 0, err
	}

	go h.processDocument(doc.ID, data)

	return doc.ID, nil
}

func (h *HttpHandler) processDocument(id uint64, data io.Reader) {
	// get document by id
	// read pdf and chunkify
	// insert chunks into database
	// insert embeddings into database
	// insert metadata into database
}
