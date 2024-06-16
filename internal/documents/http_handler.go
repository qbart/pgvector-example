package documents

import (
	"context"
	"log/slog"
	"net/http"
	"qbart/pgvector/ui"
	"qbart/pgvector/web"
	"strconv"
)

type HttpHandler struct {
	Documents Service
	Renderer  *web.Renderer
}

func (h *HttpHandler) DocumentsIndex(w http.ResponseWriter, r *http.Request) {
	page := ui.DocumentsIndexPage{
		Query:     r.URL.Query().Get("q"),
		Documents: []ui.SearchResult{},
	}
	if page.Query != "" {
		resp, err := h.Documents.SearchDocument(r.Context(), page.Query, 0.5)
		if err != nil {
			page.Error = err.Error()
		} else {
			page.Documents = make([]ui.SearchResult, len(resp))
			for i, doc := range resp {
				page.Documents[i] = ui.SearchResult{
					Title:           doc.Title,
					DocumentID:      strconv.FormatInt(doc.DocumentID, 10),
					Chunk:           strconv.FormatInt(doc.Chunk, 10),
					Content:         doc.Content,
					Page:            strconv.FormatInt(doc.Page, 10),
					Score:           strconv.FormatFloat(float64(doc.Score), 'f', 4, 32),
					AcceptableScore: doc.AcceptableScore,
				}
			}

			enhancedContext := []string{}
			for _, doc := range resp {
				if doc.AcceptableScore {
					enhancedContext = append(enhancedContext, doc.Content)
				}
			}
			answer, err := h.Documents.Ask(r.Context(), &AskInput{
				Question: page.Query,
				Context:  enhancedContext,
			})
			if err != nil {
				page.Error = err.Error()
			} else {
				md, err := h.Renderer.Markdown([]byte(answer.Answer))
				if err != nil {
					page.Error = err.Error()
				} else {
					page.Answer = string(md)
				}
			}
		}
	}
	h.Renderer.HTML(w, r, ui.DocumentsIndex(page))
}

func (h *HttpHandler) DocumentsCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(DocumentUploadLimit)
	form := r.MultipartForm
	file := form.File["file"]
	if len(file) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no file uploaded"))
	}

	fileHeader := file[0]
	src, err := fileHeader.Open()
	if err != nil {
		slog.Error("Error opening file", "message", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer src.Close()

	// indexing starts goroutine so context of request cannot be used
	// as it will be cancelled when request ends
	ctx := context.Background()
	err = h.Documents.IndexDocument(ctx, fileHeader.Filename, fileHeader.Size, src)
	if err != nil {
		slog.Error("Error saving document", "message", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/documents", http.StatusFound)
}
