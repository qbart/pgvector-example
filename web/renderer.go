package web

import (
	"bytes"
	"net/http"
	"qbart/pgvector/ui"

	"github.com/a-h/templ"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type Renderer struct{}

func (renderer *Renderer) HTML(w http.ResponseWriter, r *http.Request, view templ.Component) {
	w.Header().Set("Content-Type", "text/html")
	ui.Layout(view).Render(r.Context(), w)
}

func (renderer *Renderer) Markdown(source []byte) ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)
	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
