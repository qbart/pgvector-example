package accounts

import (
	"SoftKiwiGames/go-web-template/accounts/dto"
	"SoftKiwiGames/go-web-template/accounts/ui"
	mainui "SoftKiwiGames/go-web-template/ui"
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Plugin struct {
	Profile ProfileAction
}

type ProfileAction interface {
	GetProfile(ctx context.Context) (dto.User, error)
}

func (p *Plugin) Router() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", p.GetAccount)
	}
}

func (p *Plugin) GetAccount(w http.ResponseWriter, r *http.Request) {
	user, err := p.Profile.GetProfile(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mainui.Layout(ui.AccountPages(ui.AccountMyProfile(user))).Render(r.Context(), w)
}
