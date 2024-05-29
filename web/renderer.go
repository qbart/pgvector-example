package web

import (
	"SoftKiwiGames/go-web-template/ui"
	"net/http"

	"github.com/a-h/templ"
)

type Renderer struct {
}

func (renderer *Renderer) HTML(w http.ResponseWriter, r *http.Request, view templ.Component) {
	ui.Layout(view).Render(r.Context(), w)
}
