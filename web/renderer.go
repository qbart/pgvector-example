package web

import (
	"net/http"
	"qbart/pgvector/ui"

	"github.com/a-h/templ"
)

type Renderer struct {
}

func (renderer *Renderer) HTML(w http.ResponseWriter, r *http.Request, view templ.Component) {
	w.Header().Set("Content-Type", "text/html")
	ui.Layout(view).Render(r.Context(), w)
}
